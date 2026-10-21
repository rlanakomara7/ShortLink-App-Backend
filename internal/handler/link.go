package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	Slug        string `json:"slug"`
}

func CreateLink(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateLinkRequest

		// Read request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Original URL is required",
			})
			return
		}

		// Validate URL
		originalURL := strings.TrimSpace(req.OriginalURL)
		parsedURL, err := url.Parse(originalURL)

		if err != nil || parsedURL.Host == "" ||
			(parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Invalid URL",
			})
			return
		}

		// Create slug
		slug := req.Slug

		if slug == "" {
			randomBytes := make([]byte, 4)
			_, err := rand.Read(randomBytes)

			if err != nil {
				c.JSON(500, gin.H{
					"success": false,
					"message": "Failed to generate slug",
				})
				return
			}

			slug = hex.EncodeToString(randomBytes)
		}

		// Check slug length
		if len(slug) < 3 || len(slug) > 50 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Slug must be 3-50 characters",
			})
			return
		}

		for _, char := range slug {
			if !(char >= 'a' && char <= 'z') &&
				!(char >= 'A' && char <= 'Z') &&
				!(char >= '0' && char <= '9') && char != '-' {
				c.JSON(400, gin.H{
					"success": false,
					"message": "Slug contains invalid characters",
				})
				return
			}
		}

		switch strings.ToLower(slug) {
		case "api", "login", "register", "dashboard":
			c.JSON(400, gin.H{
				"success": false,
				"message": "Slug is reserved",
			})
			return
		}

		// Check duplicate slug
		var exists bool

		err = db.QueryRow(
			c.Request.Context(),
			"SELECT EXISTS(SELECT 1 FROM links WHERE slug = $1)",
			slug,
		).Scan(&exists)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Database error",
			})
			return
		}

		if exists {
			c.JSON(409, gin.H{
				"success": false,
				"message": "Slug already exists",
			})
			return
		}

		// Save shortlink
		userID := c.GetInt64("user_id")

		var id int64

		err = db.QueryRow(
			c.Request.Context(),
			"INSERT INTO links (user_id, original_url, slug) VALUES ($1, $2, $3) RETURNING id",
			userID,
			originalURL,
			slug,
		).Scan(&id)

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to create shortlink",
			})
			return
		}

		// Return response
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Shortlink created successfully",
			"results": gin.H{
				"id":           id,
				"original_url": originalURL,
				"slug":         slug,
				"short_url":    "http://" + c.Request.Host + "/" + slug,
			},
		})
	}
}

func GetLinks(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.GetInt64("user_id")

		rows, err := db.Query(
			c.Request.Context(),
			`SELECT id, original_url, slug
			 FROM links
			 WHERE user_id = $1 AND deleted_at IS NULL
			 ORDER BY created_at DESC, id DESC`,
			userID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get links",
			})
			return
		}
		defer rows.Close()

		links := []gin.H{}

		for rows.Next() {
			var id int64
			var originalURL string
			var slug string

			if err := rows.Scan(&id, &originalURL, &slug); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to read links",
				})
				return
			}

			links = append(links, gin.H{
				"id":           id,
				"original_url": originalURL,
				"slug":         slug,
				"short_url":    "http://" + c.Request.Host + "/" + slug,
			})
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to read links",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Links retrieved successfully",
			"results": links,
		})
	}
}
