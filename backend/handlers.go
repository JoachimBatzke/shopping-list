package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Input validation limits
const (
	maxListNameLength = 15
	maxItemNameLength = 100
	maxHexColorLength = 6
)

// List represents a shopping list
type List struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Emoji     *string   `json:"emoji"`
	HexColor  string    `json:"hex_color"`
	CreatedAt time.Time `json:"created_at"`
}

// Item represents a shopping list item
// The `json:"..."` tags tell Go how to convert to/from JSON
type Item struct {
	ID          string    `json:"id"`
	ListID      string    `json:"list_id"`
	Name        string    `json:"name"`
	Checked     bool      `json:"checked"`
	SortOrder   float64   `json:"sort_order"`
	IsSeparator bool      `json:"is_separator"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetItems handles GET /api/lists/{listId}/items - returns items for a specific list
func GetItems(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	if listID == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	// Query items for this list, ordered by sort_order then created_at
	rows, err := DB.Query(context.Background(),
		`SELECT id, list_id, name, checked, sort_order, is_separator, created_at
		 FROM items WHERE list_id = $1
		 ORDER BY sort_order ASC, created_at DESC`, listID)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.ListID, &item.Name, &item.Checked,
			&item.SortOrder, &item.IsSeparator, &item.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	if items == nil {
		items = []Item{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// CreateItem handles POST /api/lists/{listId}/items - creates a new item
func CreateItem(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	if listID == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	var input struct {
		Name        string `json:"name"`
		IsSeparator bool   `json:"is_separator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" && !input.IsSeparator {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(input.Name) > maxItemNameLength {
		http.Error(w, fmt.Sprintf("Item name must be %d characters or less", maxItemNameLength), http.StatusBadRequest)
		return
	}

	// Get max sort_order for this list to append at the end
	var maxOrder float64
	DB.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(sort_order), 0) FROM items WHERE list_id = $1",
		listID).Scan(&maxOrder)

	var item Item
	err := DB.QueryRow(context.Background(),
		`INSERT INTO items (list_id, name, is_separator, sort_order)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, list_id, name, checked, sort_order, is_separator, created_at`,
		listID, input.Name, input.IsSeparator, maxOrder+1,
	).Scan(&item.ID, &item.ListID, &item.Name, &item.Checked,
		&item.SortOrder, &item.IsSeparator, &item.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	// Track item addition for recommendations (async, don't block response)
	go TrackItemAddition(listID, input.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// UpdateItem handles PATCH /api/lists/{listId}/items/{id} - updates an item
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	id := r.PathValue("id")
	if listID == "" || id == "" {
		http.Error(w, "List ID and Item ID are required", http.StatusBadRequest)
		return
	}

	var input struct {
		Checked   *bool    `json:"checked"`
		Name      *string  `json:"name"`
		SortOrder *float64 `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate name length if provided
	if input.Name != nil && len(*input.Name) > maxItemNameLength {
		http.Error(w, fmt.Sprintf("Item name must be %d characters or less", maxItemNameLength), http.StatusBadRequest)
		return
	}

	// Build dynamic update query based on provided fields
	args := []any{}
	argNum := 1
	updates := []string{}

	if input.Checked != nil {
		updates = append(updates, fmt.Sprintf("checked = $%d", argNum))
		args = append(args, *input.Checked)
		argNum++
	}
	if input.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *input.Name)
		argNum++
	}
	if input.SortOrder != nil {
		updates = append(updates, fmt.Sprintf("sort_order = $%d", argNum))
		args = append(args, *input.SortOrder)
		argNum++
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query := "UPDATE items SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	// Verify item belongs to the specified list
	query += fmt.Sprintf(" WHERE id = $%d AND list_id = $%d", argNum, argNum+1)
	query += " RETURNING id, list_id, name, checked, sort_order, is_separator, created_at"
	args = append(args, id, listID)

	var item Item
	err := DB.QueryRow(context.Background(), query, args...).Scan(
		&item.ID, &item.ListID, &item.Name, &item.Checked,
		&item.SortOrder, &item.IsSeparator, &item.CreatedAt)

	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// ReorderItems handles PUT /api/lists/{listId}/items/reorder - reorders items
func ReorderItems(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	if listID == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	var input struct {
		ItemIDs []string `json:"item_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(input.ItemIDs) == 0 {
		http.Error(w, "item_ids is required", http.StatusBadRequest)
		return
	}

	// Update sort_order for each item based on its position in the array
	for i, itemID := range input.ItemIDs {
		_, err := DB.Exec(context.Background(),
			"UPDATE items SET sort_order = $1 WHERE id = $2 AND list_id = $3",
			float64(i+1), itemID, listID)
		if err != nil {
			http.Error(w, "Failed to reorder items", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// DeleteItem handles DELETE /api/lists/{listId}/items/{id} - deletes an item
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	id := r.PathValue("id")
	if listID == "" || id == "" {
		http.Error(w, "List ID and Item ID are required", http.StatusBadRequest)
		return
	}

	// Verify item belongs to the specified list before deleting
	result, err := DB.Exec(context.Background(),
		"DELETE FROM items WHERE id = $1 AND list_id = $2", id, listID)
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============ LIST HANDLERS ============

// GetLists handles GET /api/lists - returns all lists
func GetLists(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(context.Background(),
		"SELECT id, name, emoji, hex_color, created_at FROM lists ORDER BY created_at ASC")
	if err != nil {
		http.Error(w, "Failed to fetch lists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lists []List
	for rows.Next() {
		var list List
		err := rows.Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan list", http.StatusInternalServerError)
			return
		}
		lists = append(lists, list)
	}

	if lists == nil {
		lists = []List{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

// GetList handles GET /api/lists/{id} - returns a single list
func GetList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	var list List
	err := DB.QueryRow(context.Background(),
		"SELECT id, name, emoji, hex_color, created_at FROM lists WHERE id = $1",
		id).Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)

	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// CreateList handles POST /api/lists - creates a new list
func CreateList(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string  `json:"name"`
		Emoji    *string `json:"emoji"`
		HexColor string  `json:"hex_color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(input.Name) > maxListNameLength {
		http.Error(w, fmt.Sprintf("Name must be %d characters or less", maxListNameLength), http.StatusBadRequest)
		return
	}

	// Default color if not provided
	if input.HexColor == "" {
		input.HexColor = "42b883"
	}
	if len(input.HexColor) > maxHexColorLength {
		http.Error(w, "Invalid hex color", http.StatusBadRequest)
		return
	}

	var list List
	err := DB.QueryRow(context.Background(),
		`INSERT INTO lists (name, emoji, hex_color)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, emoji, hex_color, created_at`,
		input.Name, input.Emoji, input.HexColor,
	).Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to create list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(list)
}

// UpdateList handles PATCH /api/lists/{id} - updates a list
func UpdateList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	var input struct {
		Name     *string `json:"name"`
		Emoji    *string `json:"emoji"`
		HexColor *string `json:"hex_color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input lengths
	if input.Name != nil && len(*input.Name) > maxListNameLength {
		http.Error(w, fmt.Sprintf("Name must be %d characters or less", maxListNameLength), http.StatusBadRequest)
		return
	}
	if input.HexColor != nil && len(*input.HexColor) > maxHexColorLength {
		http.Error(w, "Invalid hex color", http.StatusBadRequest)
		return
	}

	// Fetch current list first
	var list List
	err := DB.QueryRow(context.Background(),
		"SELECT id, name, emoji, hex_color, created_at FROM lists WHERE id = $1",
		id).Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	// Apply updates
	if input.Name != nil {
		list.Name = *input.Name
	}
	if input.Emoji != nil {
		list.Emoji = input.Emoji
	}
	if input.HexColor != nil {
		list.HexColor = *input.HexColor
	}

	// Save changes
	err = DB.QueryRow(context.Background(),
		`UPDATE lists SET name = $1, emoji = $2, hex_color = $3 WHERE id = $4
		 RETURNING id, name, emoji, hex_color, created_at`,
		list.Name, list.Emoji, list.HexColor, id,
	).Scan(&list.ID, &list.Name, &list.Emoji, &list.HexColor, &list.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to update list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// DeleteList handles DELETE /api/lists/{id} - deletes a list and its items
func DeleteList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	result, err := DB.Exec(context.Background(),
		"DELETE FROM lists WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Failed to delete list", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============ RECOMMENDATIONS HANDLERS ============

// ItemHistory represents an item's addition history for recommendations
type ItemHistory struct {
	ID              string    `json:"id"`
	ListID          string    `json:"list_id"`
	ItemName        string    `json:"item_name"`
	AddedCount      int       `json:"added_count"`
	LastAddedAt     time.Time `json:"last_added_at"`
	AvgDaysBetween  float64   `json:"avg_days_between"`
	Dismissed       bool      `json:"dismissed"`
}

// Recommendation represents a suggested item
type Recommendation struct {
	Name    string  `json:"name"`
	Urgency float64 `json:"urgency"`
}

// GetRecommendations returns item suggestions based on addition history
func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	if listID == "" {
		http.Error(w, "List ID is required", http.StatusBadRequest)
		return
	}

	// Check if item_history table exists by querying it
	// If it doesn't exist, return empty array
	rows, err := DB.Query(context.Background(),
		`SELECT item_name,
			CASE
				WHEN avg_days_between > 0 THEN
					EXTRACT(EPOCH FROM (NOW() - last_added_at)) / 86400 / avg_days_between
				ELSE 0.5
			END as urgency
		FROM item_history
		WHERE list_id = $1
			AND dismissed = false
			AND added_count >= 2
			AND item_name NOT IN (SELECT name FROM items WHERE list_id = $1)
		ORDER BY urgency DESC
		LIMIT 10`, listID)
	if err != nil {
		// Table might not exist - return empty recommendations
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	defer rows.Close()

	var recs []Recommendation
	for rows.Next() {
		var rec Recommendation
		if err := rows.Scan(&rec.Name, &rec.Urgency); err != nil {
			continue
		}
		if rec.Urgency >= 0.5 { // Only suggest items with decent urgency
			recs = append(recs, rec)
		}
	}

	if recs == nil {
		recs = []Recommendation{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

// DismissRecommendation dismisses a recommendation
func DismissRecommendation(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("listId")
	name := r.PathValue("name")
	if listID == "" || name == "" {
		http.Error(w, "List ID and item name are required", http.StatusBadRequest)
		return
	}

	_, err := DB.Exec(context.Background(),
		`UPDATE item_history SET dismissed = true
		WHERE list_id = $1 AND item_name = $2`,
		listID, name)
	if err != nil {
		http.Error(w, "Failed to dismiss recommendation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// TrackItemAddition updates item history when an item is added
func TrackItemAddition(listID, itemName string) {
	// Try to update existing record
	result, err := DB.Exec(context.Background(),
		`UPDATE item_history
		SET added_count = added_count + 1,
			avg_days_between = CASE
				WHEN added_count > 0 THEN
					(avg_days_between * added_count + EXTRACT(EPOCH FROM (NOW() - last_added_at)) / 86400) / (added_count + 1)
				ELSE 7
			END,
			last_added_at = NOW(),
			dismissed = false
		WHERE list_id = $1 AND item_name = $2`,
		listID, itemName)

	if err != nil || result.RowsAffected() == 0 {
		// Insert new record
		DB.Exec(context.Background(),
			`INSERT INTO item_history (list_id, item_name, added_count, last_added_at, avg_days_between, dismissed)
			VALUES ($1, $2, 1, NOW(), 7, false)
			ON CONFLICT DO NOTHING`,
			listID, itemName)
	}
}
