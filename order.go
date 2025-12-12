
package main

import (
	"database/sql"
	"log"
	"time"
)

// Order represents an order in the system
type Order struct {
	ID        int         `json:"id"`
	UserID    int         `json:"userId"`
	CreatedAt time.Time   `json:"createdAt"`
	Items     []OrderItem `json:"items"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"orderId"`
	ProductID int     `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"` // Price at the time of purchase
}

// CreateOrder creates a new order from a cart
func CreateOrder(db *sql.DB, cartID, userID int) (*Order, error) {
	log.Printf("Creating order for cartID: %d, userID: %d", cartID, userID)
	cart, err := GetCart(db, cartID)
	if err != nil {
		log.Printf("Error getting cart: %v", err)
		return nil, err
	}

	if len(cart.Items) == 0 {
		log.Println("Cannot create an empty order")
		return nil, nil // Cannot create an empty order
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return nil, err
	}

	// Create the order
	res, err := tx.Exec("INSERT INTO orders (user_id, created_at) VALUES (?, ?)", userID, time.Now())
	if err != nil {
		tx.Rollback()
		log.Printf("Error creating order: %v", err)
		return nil, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Printf("Error getting last insert ID: %v", err)
		return nil, err
	}

	// Create the order items
	for _, item := range cart.Items {
		_, err := tx.Exec("INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)",
			orderID, item.ProductID, item.Quantity, item.Product.Price)
		if err != nil {
			tx.Rollback()
			log.Printf("Error creating order item: %v", err)
			return nil, err
		}
	}

	// Clear the cart
	_, err = tx.Exec("DELETE FROM cart_items WHERE cart_id = ?", cartID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error clearing cart: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	log.Printf("Order created successfully with ID: %d", orderID)
	return &Order{ID: int(orderID), UserID: userID}, nil
}

// GetOrdersByUserID retrieves all orders for a given user
func GetOrdersByUserID(db *sql.DB, userID int) ([]Order, error) {
	log.Printf("Getting orders for userID: %d", userID)
	rows, err := db.Query("SELECT id, created_at FROM orders WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		order.UserID = userID
		if err := rows.Scan(&order.ID, &order.CreatedAt); err != nil {
			log.Printf("Error scanning order: %v", err)
			return nil, err
		}

		// Get order items
		itemRows, err := db.Query("SELECT id, product_id, quantity, price FROM order_items WHERE order_id = ?", order.ID)
		if err != nil {
			log.Printf("Error getting order items: %v", err)
			return nil, err
		}

		for itemRows.Next() {
			var item OrderItem
			item.OrderID = order.ID
			if err := itemRows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
				itemRows.Close()
				log.Printf("Error scanning order item: %v", err)
				return nil, err
			}
			order.Items = append(order.Items, item)
		}
		itemRows.Close()

		orders = append(orders, order)
	}

	log.Printf("Found %d orders for userID: %d", len(orders), userID)
	return orders, nil
}
