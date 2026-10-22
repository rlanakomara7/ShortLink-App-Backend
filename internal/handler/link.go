package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	Slug        string `json:"slug"`
}

// CreateLink godoc
// @Summary Create a shortlink
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateLinkRequest true "Shortlink data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/links [post]
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

// GetLinks godoc
// @Summary Get current user's shortlinks
// @Tags Shortlinks
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/links [get]
func GetLinks(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID := c.GetInt64("user_id")

		rows, err := db.Query(
			c.Request.Context(),
			`SELECT id, original_url, slug, created_at
     		FROM links
     		WHERE user_id = $1 AND deleted_at IS NULL
     		ORDER BY created_at DESC`,
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
			var createdAt time.Time

			if err := rows.Scan(&id, &originalURL, &slug, &createdAt); err != nil {
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
				"created_at":   createdAt,
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

// DeleteLink godoc
// @Summary Delete a shortlink
// @Tags Shortlinks
// @Produce json
// @Security BearerAuth
// @Param id path int true "Link ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/links/{id} [delete]
func DeleteLink(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Convert link ID to integer
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid link ID",
			})
			return
		}

		userID := c.GetInt64("user_id")

		result, err := db.Exec(
			c.Request.Context(),
			`UPDATE links
			 SET deleted_at = NOW()
			 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
			id, userID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete shortlink",
			})
			return
		}

		if result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Link not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Shortlink deleted successfully",
		})
	}
}

// RedirectLink godoc
// @Summary Redirect to the original URL
// @Tags Shortlinks
// @Param slug path string true "Shortlink slug"
// @Success 301 {string} string "Moved Permanently"
// @Failure 404 {object} map[string]interface{}
// @Router /{slug} [get]
func RedirectLink(db *pgxpool.Pool, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		ctx := c.Request.Context()
		cacheKey := "shortlink:" + slug

		// Check Redis cache
		originalURL, err := redisClient.Get(ctx, cacheKey).Result()

		if err == nil {
			c.Redirect(http.StatusMovedPermanently, originalURL)
			return
		}

		if err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Cache error",
			})
			return
		}

		// Get original URL from PostgreSQL
		err = db.QueryRow(
			ctx,
			"SELECT original_url FROM links WHERE slug = $1 AND deleted_at IS NULL",
			slug,
		).Scan(&originalURL)

		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Shortlink not found",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Database error",
			})
			return
		}

		redisClient.Set(ctx, cacheKey, originalURL, 10*time.Minute)

		c.Redirect(http.StatusMovedPermanently, originalURL)
	}
}
