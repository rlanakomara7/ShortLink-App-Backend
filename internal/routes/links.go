package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/rlanakomara7/ShortLink-App-Backend/internal/handler"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/middleware"
)

func LinkRoutes(router *gin.Engine, db *pgxpool.Pool, jwtSecret string, redisClient *redis.Client) {

	//public route
	router.GET("/:slug", handler.RedirectLink(db, redisClient))

	//protected route
	protected := router.Group("/api")
	protected.Use(middleware.Auth(jwtSecret, redisClient))

	protected.POST("/links", handler.CreateLink(db))
	protected.GET("/links", handler.GetLinks(db))
	protected.DELETE("/links/:id", handler.DeleteLink(db))
	protected.GET("/profile", handler.GetProfile(db))

}
