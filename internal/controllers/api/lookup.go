package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"geodbsvc/internal/response"
)

// Lookup получает информацию об IP
func (c *Controller) Lookup() gin.HandlerFunc {
	db := c.srv.GetDB()
	logger := c.srv.GetLogger()

	validate := validator.New()

	return func(ctx *gin.Context) {
		ip := ctx.Query("ip")

		if err := validate.Var(ip, "required,ip"); err != nil {
			logger.Errorf("Error: %v", err)
			response.SendError(ctx, "Переданы неверные параметры")
			return
		}

		rec, err := db.Lookup(ip)
		if err != nil {
			logger.Errorf("Error: %v", err)
			response.SendError(ctx, "Возникла ошибка при получении информации")
			return
		}

		response.SendSuccess(ctx, "Информация получена успешно", rec)
	}
}
