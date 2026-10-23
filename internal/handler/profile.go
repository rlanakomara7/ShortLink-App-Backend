package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileData struct {
	Email     string `json:"email" example:"rama@gmail.com"`
	CreatedAt string `json:"created_at" example:"2026-10-22"`
}

type ProfileResponse struct {
	Success bool        `json:"success" example:"true"`
	Results ProfileData `json:"results"`
}

// GetProfile Godoc
// @Summary Get user profile
// @Description Get the authenticated user's email and registration date
// @Tags Profile
// @Produce json
// @Tags Profile
// @Produce json
// @Success 200 {object} ProfileResponse "Profile retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Failed to load profile"
// @Router /api/profile [get]
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
