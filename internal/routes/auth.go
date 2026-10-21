package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rlanakomara7/ShortLink-App-Backend/internal/handler"
)

func AuthRoutes(router *gin.Engine, db *pgxpool.Pool, jwtSecret string) {
	router.POST("/api/register", handler.Register(db))
	router.POST("/api/login", handler.Login(db, jwtSecret))
}
