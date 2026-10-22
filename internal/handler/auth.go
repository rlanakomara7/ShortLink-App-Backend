package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register godoc
// @Summary Register a new user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/register [post]
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

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary Login and get JWT access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/login [post]
func Login(db *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest

		// Validate request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid email or password",
			})
			return
		}

		email := strings.ToLower(strings.TrimSpace(req.Email))

		// Find user by email
		var id int64
		var passwordHash string

		err := db.QueryRow(
			c.Request.Context(),
			"SELECT id, password_hash FROM users WHERE LOWER(email) = $1",
			email,
		).Scan(&id, &passwordHash)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid email or password",
			})
			return
		}

		// Verify password
		err = bcrypt.CompareHashAndPassword(
			[]byte(passwordHash),
			[]byte(req.Password),
		)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid email or password",
			})
			return
		}

		// Generate JWT access token
		claims := jwt.MapClaims{
			"user_id": id,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to generate token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Login successful",
			"results": gin.H{
				"access_token": tokenString,
			},
		})
	}
}
