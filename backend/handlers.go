package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// UpdateItem handles PATCH /api/items/{id} - updates an item
func UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
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
	query += fmt.Sprintf(" WHERE id = $%d", argNum)
	query += " RETURNING id, list_id, name, checked, sort_order, is_separator, created_at"
	args = append(args, id)

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

// DeleteItem handles DELETE /api/items/{id} - deletes an item
func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	result, err := DB.Exec(context.Background(),
		"DELETE FROM items WHERE id = $1", id)
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

	// Default color if not provided
	if input.HexColor == "" {
		input.HexColor = "42b883"
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
