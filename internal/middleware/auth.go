package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func Auth(secret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read Bearer token
		header := c.GetHeader("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
			})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		// Validate JWT
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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok || userID < 1 || userID != float64(int64(userID)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		expiresAt, err := claims.GetExpirationTime()
		if err != nil || expiresAt == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check token blacklist
		hash := sha256.Sum256([]byte(tokenString))
		tokenHash := hex.EncodeToString(hash[:])
		cacheKey := "blacklist:" + tokenHash

		blocked, err := redisClient.Exists(
			c.Request.Context(),
			cacheKey,
		).Result()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Authentication service unavailable",
			})
			return
		}

		if blocked > 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token has been revoked",
			})
			return
		}

		c.Set("user_id", int64(userID))
		c.Set("access_token", tokenString)
		c.Set("token_expires_at", expiresAt.Time)

		c.Next()
	}
}
