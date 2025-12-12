package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database and creates tables if they don't exist
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	// Enable foreign key support
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Create categories table
	statement, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create products table
	statement, err = db.Prepare(`
        CREATE TABLE IF NOT EXISTS products (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            price REAL NOT NULL,
            image_url TEXT,
			category_id INTEGER,
			FOREIGN KEY(category_id) REFERENCES categories(id)
        )
    `)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create reviews table
	statement, err = db.Prepare(`
		CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			rating INTEGER NOT NULL,
			comment TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY(product_id) REFERENCES products(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create carts table
	statement, err = db.Prepare(`
		CREATE TABLE IF NOT EXISTS carts (
			id INTEGER PRIMARY KEY AUTOINCREMENT
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create cart_items table
	statement, err = db.Prepare(`
		CREATE TABLE IF NOT EXISTS cart_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cart_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			FOREIGN KEY(cart_id) REFERENCES carts(id),
			FOREIGN KEY(product_id) REFERENCES products(id)
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create users table
	statement, err = db.Prepare(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create orders table
	statement, err = db.Prepare(`
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Create order_items table
	statement, err = db.Prepare(`
		CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			price REAL NOT NULL,
			FOREIGN KEY(order_id) REFERENCES orders(id),
			FOREIGN KEY(product_id) REFERENCES products(id)
		)
	`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	// Insert some sample data
	// In a real application, you would have a separate seeding process
	count := 0
	db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if count == 0 {
		db.Exec(`INSERT INTO categories (name) VALUES ('Laptops'), ('Smartphones'), ('Books'), ('T-Shirts'), ('Headphones')`)
	}

	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if count == 0 {
		db.Exec(`INSERT INTO products (name, description, price, image_url, category_id) VALUES
            ('MacBook Pro', 'The latest MacBook Pro with M3 chip.', 2500.00, 'https://placeimg.com/640/480/tech', 1),
            ('Dell XPS 15', 'A powerful and stylish Windows laptop.', 2000.00, 'https://placeimg.com/640/480/tech?2', 1),
            ('iPhone 15 Pro', 'The latest iPhone with A17 Pro chip.', 1200.00, 'https://placeimg.com/640/480/tech?3', 2),
            ('Samsung Galaxy S24', 'The latest Samsung phone with Galaxy AI.', 1100.00, 'https://placeimg.com/640/480/tech?4', 2),
            ('The Pragmatic Programmer', 'Your journey to mastery, 20th Anniversary Edition.', 50.00, 'https://placeimg.com/640/480/arch', 3),
            ('Clean Code', 'A Handbook of Agile Software Craftsmanship.', 45.00, 'https://placeimg.com/640/480/arch?2', 3),
            ('Go-Commerce T-Shirt', 'A comfortable and stylish t-shirt for Go developers.', 30.00, 'https://placeimg.com/640/480/people', 4),
            ('Fiber T-Shirt', 'Show your love for the Fiber framework.', 30.00, 'https://placeimg.com/640/480/people?2', 4),
            ('Sony WH-1000XM5', 'Industry-leading noise canceling headphones.', 400.00, 'https://placeimg.com/640/480/tech?5', 5),
            ('Bose QuietComfort Ultra', 'The next generation of noise-cancelling headphones.', 430.00, 'https://placeimg.com/640/480/tech?6', 5)
		`)
	}

	return db, nil
}