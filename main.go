package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
)

func main() {
	// Initialize database
	db, err := InitDB("./database/store.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize session store
	store := session.New(session.Config{
		Storage: sqlite3.New(sqlite3.Config{
			Database: "./database/store.db",
			Table:    "sessions",
		}),
		Expiration: 24 * time.Hour,
	})

	app := fiber.New()

	app.Static("/", "./public")

	api := app.Group("/api")

	// Products endpoints
	api.Get("/products", func(c *fiber.Ctx) error {
		searchTerm := c.Query("search")
		categoryID := c.Query("category")
		products, err := GetProducts(db, searchTerm, categoryID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(products)
	})

	api.Get("/products/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID")
		}

		product, err := GetProduct(db, id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if product == nil {
			return fiber.NewError(fiber.StatusNotFound, "Product not found")
		}

		return c.JSON(product)
	})

	// Reviews endpoints
	api.Get("/products/:id/reviews", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID")
		}

		reviews, err := GetReviewsByProductID(db, id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(reviews)
	})

	type CreateReviewRequest struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}

	api.Post("/products/:id/reviews", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID")
		}

		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		userID, ok := sess.Get("userID").(int)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "Not logged in")
		}

		var req CreateReviewRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		review, err := CreateReview(db, id, userID, req.Rating, req.Comment)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(review)
	})

	// Categories endpoint
	api.Get("/categories", func(c *fiber.Ctx) error {
		categories, err := GetCategories(db)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(categories)
	})

	// Cart endpoints
	api.Post("/cart", func(c *fiber.Ctx) error {
		res, err := db.Exec("INSERT INTO carts DEFAULT VALUES")
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		id, err := res.LastInsertId()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"id": id})
	})

	api.Get("/cart/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid cart ID")
		}

		cart, err := GetCart(db, id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(cart)
	})

	type AddToCartRequest struct {
		ProductID int `json:"productId"`
		Quantity  int `json:"quantity"`
	}

	api.Post("/cart/:id/items", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid cart ID")
		}

		var req AddToCartRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if err := AddItemToCart(db, id, req.ProductID, req.Quantity); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	// Auth endpoints
	api.Get("/me", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		userID, ok := sess.Get("userID").(int)
		if !ok {
			return c.JSON(fiber.Map{"loggedIn": false})
		}

		return c.JSON(fiber.Map{"loggedIn": true, "userID": userID})
	})

	type AuthRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	api.Post("/register", func(c *fiber.Ctx) error {
		var req AuthRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		user, err := CreateUser(db, req.Username, req.Password)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(user)
	})

	api.Post("/login", func(c *fiber.Ctx) error {
		var req AuthRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		user, err := GetUserByUsername(db, req.Username)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if user == nil || !CheckPasswordHash(req.Password, user.Password) {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
		}

		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		sess.Set("userID", user.ID)
		if err := sess.Save(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{"message": "Login successful"})
	})

	api.Post("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if err := sess.Destroy(); err != nil {
		    return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Orders endpoints
	type CreateOrderRequest struct {
		CartID int `json:"cartId"`
	}

	api.Post("/orders", func(c *fiber.Ctx) error {
		log.Println("Received request to create order")
		sess, err := store.Get(c)
		if err != nil {
			log.Printf("Error getting session: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		userID, ok := sess.Get("userID").(int)
		if !ok {
			log.Println("User not logged in")
			return fiber.NewError(fiber.StatusUnauthorized, "Not logged in")
		}

		var req CreateOrderRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("Error parsing request body: %v", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		order, err := CreateOrder(db, req.CartID, userID)
		if err != nil {
			log.Printf("Error creating order: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if order == nil {
			log.Println("Cannot create an empty order")
			return fiber.NewError(fiber.StatusBadRequest, "Cannot create an empty order")
		}

		log.Printf("Order created successfully: %v", order)
		return c.JSON(order)
	})

	api.Get("/orders", func(c *fiber.Ctx) error {
		log.Println("Received request to get orders")
		sess, err := store.Get(c)
		if err != nil {
			log.Printf("Error getting session: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		userID, ok := sess.Get("userID").(int)
		if !ok {
			log.Println("User not logged in")
			return fiber.NewError(fiber.StatusUnauthorized, "Not logged in")
		}

		orders, err := GetOrdersByUserID(db, userID)
		if err != nil {
			log.Printf("Error getting orders by user ID: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		log.Printf("Returning %d orders for user %d", len(orders), userID)
		return c.JSON(orders)
	})

	app.Listen(":3000")
}