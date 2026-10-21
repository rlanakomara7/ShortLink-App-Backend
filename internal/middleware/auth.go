package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		tokenString := strings.TrimPrefix(
			c.GetHeader("Authorization"), "Bearer ",
		)

		// Validate token
		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		// Get user ID from token
		claims := token.Claims.(jwt.MapClaims)
		userID, ok := claims["user_id"].(float64)

		if !ok || userID < 1 || userID != float64(int64(userID)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		c.Set("user_id", int64(userID))
		c.Next()
	}
}
