package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Item represents a shopping list item
// The `json:"..."` tags tell Go how to convert to/from JSON
type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Checked   bool      `json:"checked"`
	CreatedAt time.Time `json:"created_at"`
}

// GetItems handles GET /api/items - returns all items
func GetItems(w http.ResponseWriter, r *http.Request) {
	// Query the database for all items, newest first
	rows, err := DB.Query(context.Background(),
		"SELECT id, name, checked, created_at FROM items ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Always close rows when done!

	// Collect all items into a slice (Go's version of an array)
	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Checked, &item.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	// Return empty array instead of null if no items
	if items == nil {
		items = []Item{}
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// CreateItem handles POST /api/items - creates a new item
func CreateItem(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body from the request
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate: name can't be empty
	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Insert into database and get the created item back
	var item Item
	err := DB.QueryRow(context.Background(),
		`INSERT INTO items (name) VALUES ($1)
		 RETURNING id, name, checked, created_at`,
		input.Name,
	).Scan(&item.ID, &item.Name, &item.Checked, &item.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	// Return the created item with 201 Created status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// UpdateItem handles PATCH /api/items/{id} - updates an item (toggle checked)
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Get the item ID from the URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	// Parse the JSON body
	var input struct {
		Checked *bool `json:"checked"` // Pointer so we can detect if it was provided
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Checked == nil {
		http.Error(w, "checked field is required", http.StatusBadRequest)
		return
	}

	// Update in database
	var item Item
	err := DB.QueryRow(context.Background(),
		`UPDATE items SET checked = $1 WHERE id = $2
		 RETURNING id, name, checked, created_at`,
		*input.Checked, id,
	).Scan(&item.ID, &item.Name, &item.Checked, &item.CreatedAt)

	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteItem handles DELETE /api/items/{id} - deletes an item
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	// Delete from database
	result, err := DB.Exec(context.Background(),
		"DELETE FROM items WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	// Check if anything was actually deleted
	if result.RowsAffected() == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Return 204 No Content (success, but nothing to return)
	w.WriteHeader(http.StatusNoContent)
}
