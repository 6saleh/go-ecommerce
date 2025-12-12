
package main

import (
	"database/sql"
	"time"
)

// Review represents a user review for a product
type Review struct {
	ID        int       `json:"id"`
	ProductID int       `json:"productId"`
	UserID    int       `json:"userId"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetReviewsByProductID retrieves all reviews for a given product
func GetReviewsByProductID(db *sql.DB, productID int) ([]Review, error) {
	rows, err := db.Query("SELECT id, user_id, rating, comment, created_at FROM reviews WHERE product_id = ? ORDER BY created_at DESC", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		r.ProductID = productID
		if err := rows.Scan(&r.ID, &r.UserID, &r.Rating, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

// CreateReview creates a new review for a product
func CreateReview(db *sql.DB, productID, userID, rating int, comment string) (*Review, error) {
	res, err := db.Exec("INSERT INTO reviews (product_id, user_id, rating, comment, created_at) VALUES (?, ?, ?, ?, ?)", productID, userID, rating, comment, time.Now())
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Review{ID: int(id), ProductID: productID, UserID: userID, Rating: rating, Comment: comment}, nil
}
