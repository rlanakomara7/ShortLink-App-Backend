package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rlanakomara7/ShortLink-App-Backend/internal/handler"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/middleware"
)

func LinkRoutes(router *gin.Engine, db *pgxpool.Pool, jwtSecret string) {
	protected := router.Group("/api")
	protected.Use(middleware.Auth(jwtSecret))

	protected.POST("/links", handler.CreateLink(db))
}
