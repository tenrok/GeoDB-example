package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"geodb-example/internal/response"
)

// Lookup получает информацию об IP
func (c *Controller) Lookup() gin.HandlerFunc {
	db := c.srv.DB()
	logger := c.srv.Logger()

	validate := validator.New()

	return func(ctx *gin.Context) {
		ip := ctx.Query("ip")

		if err := validate.Var(ip, "required,ip"); err != nil {
			logger.Errorf("Error: %v", err)
			response.SendErrorf(ctx, "Переданы неверные параметры: %v", err)
			return
		}

		rec, err := db.Lookup(ip)
		if err != nil {
			logger.Errorf("Error: %v", err)
			response.SendErrorf(ctx, "Возникла ошибка при получении информации: %v", err)
			return
		}

		response.SendSuccess(ctx, "Информация получена успешно", rec)
	}
}
