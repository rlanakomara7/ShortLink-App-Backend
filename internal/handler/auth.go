package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func Register(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest

		// Validate request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid email or password",
			})
			return
		}

		// Normalize email
		email := strings.ToLower(strings.TrimSpace(req.Email))

		// Check duplicate email
		var exists bool
		err := db.QueryRow(
			c.Request.Context(),
			"SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email) = $1)",
			email,
		).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Database error",
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "Email already registered",
			})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword(
			[]byte(req.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to process password",
			})
			return
		}

		// Save user
		var id int64
		err = db.QueryRow(
			c.Request.Context(),
			"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
			email, string(hash),
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create account",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Registration successful",
			"results": gin.H{
				"id":    id,
				"email": email,
			},
		})
	}
}
