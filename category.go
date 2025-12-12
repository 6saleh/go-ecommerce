
package main

import "database/sql"

// Category represents a product category
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetCategories retrieves all categories from the database
func GetCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}
