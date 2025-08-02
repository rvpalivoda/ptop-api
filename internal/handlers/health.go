package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Health godoc
// @Summary Проверка состояния сервиса
// @Tags health
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 503 {object} ErrorResponse
// @Router /health [get]
func Health(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "db error"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "db down"})
			return
		}
		c.JSON(http.StatusOK, StatusResponse{Status: "ok"})
	}
}
