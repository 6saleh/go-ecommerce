package main

import (
	"database/sql"
	"strings"
)

// Product represents a product in the store
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"imageUrl"`
	CategoryID  int     `json:"categoryId"`
}

// GetProducts retrieves all products from the database
func GetProducts(db *sql.DB, searchTerm, categoryID string) ([]Product, error) {
	query := "SELECT id, name, description, price, image_url, category_id FROM products"
	var args []interface{}
	var whereClauses []string

	if searchTerm != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR description LIKE ?)")
		args = append(args, "%"+searchTerm+"%", "%"+searchTerm+"%")
	}

	if categoryID != "" {
		whereClauses = append(whereClauses, "category_id = ?")
		args = append(args, categoryID)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.CategoryID); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetProduct retrieves a single product from the database
func GetProduct(db *sql.DB, id int) (*Product, error) {
	row := db.QueryRow("SELECT id, name, description, price, image_url, category_id FROM products WHERE id = ?", id)

	var p Product
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.CategoryID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}

	return &p, nil
}