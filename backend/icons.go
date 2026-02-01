package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// GetListIcon generates a PNG icon with the list's emoji on its color background
// GET /api/lists/{listId}/icon/{size}.png
func GetListIcon(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	sizeStr := r.PathValue("size")

	if listID == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	// Parse size from path (e.g., "192.png" -> 192)
	sizeStr = strings.TrimSuffix(sizeStr, ".png")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 16 || size > 1024 {
		http.Error(w, "Invalid size (16-1024)", http.StatusBadRequest)
		return
	}

	// Fetch list from database
	var list List
	err = DB.QueryRow(context.Background(),
		"SELECT id, name, emoji, hex_color, created_at FROM lists WHERE id = $1",
		listID).Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	// Parse hex color
	hexColor := list.HexColor
	if len(hexColor) != 6 {
		hexColor = "333333" // Default gray
	}
	bgColor := parseHexColor(hexColor)

	// Create background image
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// If list has an emoji, fetch and composite it
	if list.Emoji != nil && *list.Emoji != "" {
		emojiImg, err := fetchEmojiImage(*list.Emoji)
		if err == nil && emojiImg != nil {
			// Calculate emoji size and position (80% of icon size, centered)
			emojiSize := int(float64(size) * 0.7)
			offset := (size - emojiSize) / 2

			// Scale emoji to fit
			scaled := image.NewRGBA(image.Rect(0, 0, emojiSize, emojiSize))
			xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), emojiImg, emojiImg.Bounds(), xdraw.Over, nil)

			// Composite onto background
			draw.Draw(img, image.Rect(offset, offset, offset+emojiSize, offset+emojiSize),
				scaled, image.Point{}, draw.Over)
		}
	}

	// Set headers
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache 24 hours

	// Encode and send
	png.Encode(w, img)
}

// GetListManifest returns a dynamic PWA manifest for a specific list
// GET /api/lists/{listId}/manifest.webmanifest
func GetListManifest(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	if listID == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	// Fetch list from database
	var list List
	err := DB.QueryRow(context.Background(),
		"SELECT id, name, emoji, hex_color, created_at FROM lists WHERE id = $1",
		listID).Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	// Get the API base URL for icon paths
	scheme := "https"
	if r.TLS == nil && !strings.Contains(r.Host, "railway") {
		scheme = "http"
	}
	apiBaseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	// Get frontend URL from Referer header or CORS_ORIGIN env var
	// This is needed because start_url must be absolute for iOS
	frontendURL := os.Getenv("CORS_ORIGIN")
	if frontendURL == "" || frontendURL == "*" {
		// Try to get from Referer header
		if referer := r.Header.Get("Referer"); referer != "" {
			if u, err := url.Parse(referer); err == nil {
				frontendURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			}
		}
	}
	// Fallback to localhost for development
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Build manifest with absolute URLs for iOS compatibility
	manifest := map[string]interface{}{
		"name":             fmt.Sprintf("JORLIST - %s", list.Name),
		"short_name":       list.Name,
		"description":      "Share a list with friends",
		"start_url":        fmt.Sprintf("%s/list/%s", frontendURL, listID),
		"scope":            fmt.Sprintf("%s/", frontendURL),
		"display":          "standalone",
		"theme_color":      fmt.Sprintf("#%s", list.HexColor),
		"background_color": fmt.Sprintf("#%s", list.HexColor),
		"icons": []map[string]string{
			{
				"src":     fmt.Sprintf("%s/api/lists/%s/icon/192.png", apiBaseURL, listID),
				"sizes":   "192x192",
				"type":    "image/png",
				"purpose": "any maskable",
			},
			{
				"src":     fmt.Sprintf("%s/api/lists/%s/icon/512.png", apiBaseURL, listID),
				"sizes":   "512x512",
				"type":    "image/png",
				"purpose": "any maskable",
			},
		},
	}

	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "no-cache") // Always fetch fresh manifest
	json.NewEncoder(w).Encode(manifest)
}

// parseHexColor converts a 6-character hex string to color.RGBA
func parseHexColor(hex string) color.RGBA {
	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{r, g, b, 255}
}

// fetchEmojiImage fetches an emoji PNG from the emoji CDN
func fetchEmojiImage(emoji string) (image.Image, error) {
	// URL encode the emoji
	encodedEmoji := url.PathEscape(emoji)
	emojiURL := fmt.Sprintf("https://emojicdn.elk.sh/%s?style=apple", encodedEmoji)

	resp, err := http.Get(emojiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch emoji: %d", resp.StatusCode)
	}

	// Read body with size limit (1MB max for emoji)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}

	// Decode PNG
	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return img, nil
}
