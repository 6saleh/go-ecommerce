
package main

import "database/sql"

// Cart represents a shopping cart
type Cart struct {
	ID    int        `json:"id"`
	Items []CartItem `json:"items"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ID        int     `json:"id"`
	CartID    int     `json:"cartId"`
	ProductID int     `json:"productId"`
	Quantity  int     `json:"quantity"`
	Product   Product `json:"product"`
}

// GetCart retrieves a cart and its items from the database
func GetCart(db *sql.DB, id int) (*Cart, error) {
	cart := &Cart{ID: id}

	rows, err := db.Query(`
		SELECT ci.id, ci.quantity, p.id, p.name, p.description, p.price, p.image_url
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.Quantity, &item.Product.ID, &item.Product.Name, &item.Product.Description, &item.Product.Price, &item.Product.ImageURL); err != nil {
			return nil, err
		}
		item.CartID = id
		item.ProductID = item.Product.ID
		cart.Items = append(cart.Items, item)
	}

	return cart, nil
}

// AddItemToCart adds an item to a cart in the database
func AddItemToCart(db *sql.DB, cartID, productID, quantity int) error {
	var existingQuantity int
	err := db.QueryRow("SELECT quantity FROM cart_items WHERE cart_id = ? AND product_id = ?", cartID, productID).Scan(&existingQuantity)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO cart_items (cart_id, product_id, quantity) VALUES (?, ?, ?)", cartID, productID, quantity)
	} else {
		_, err = db.Exec("UPDATE cart_items SET quantity = ? WHERE cart_id = ? AND product_id = ?", existingQuantity+quantity, cartID, productID)
	}

	return err
}
