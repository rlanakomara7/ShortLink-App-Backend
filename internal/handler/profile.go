package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetProfile(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")

		var email string
		var createdAt string

		err := db.QueryRow(
			c.Request.Context(),
			`SELECT email, TO_CHAR(created_at, 'YYYY-MM-DD')
			 FROM users WHERE id = $1`,
			userID,
		).Scan(&email, &createdAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to load profile",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"results": gin.H{
				"email":      email,
				"created_at": createdAt,
			},
		})
	}
}
