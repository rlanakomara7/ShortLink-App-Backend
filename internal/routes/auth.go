package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/rlanakomara7/ShortLink-App-Backend/internal/handler"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/middleware"
)

func AuthRoutes(router *gin.Engine, db *pgxpool.Pool, jwtSecret string, redisClient *redis.Client) {
	router.POST("/api/register", handler.Register(db))
	router.POST("/api/login", handler.Login(db, jwtSecret))
	router.POST("/api/logout", middleware.Auth(jwtSecret, redisClient), handler.Logout(redisClient))
}
