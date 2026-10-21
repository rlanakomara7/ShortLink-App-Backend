package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(router *gin.Engine, db *pgxpool.Pool, jwtSecret string) {
	AuthRoutes(router, db, jwtSecret)
	LinkRoutes(router, db, jwtSecret)
}
