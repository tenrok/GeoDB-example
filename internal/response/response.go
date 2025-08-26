package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseCode = uint

// Коды сообщений, отправляемых браузеру
const (
	CodeSuccess ResponseCode = 0 // Сообщение об удачном выполнении
	CodeError   ResponseCode = 1 // Возникла какая-то ошибка
)

type Response struct {
	Code   ResponseCode `json:"code"`
	Msg    string       `json:"msg,omitempty"`
	Result any          `json:"result,omitempty"`
}

// SendSuccess
func SendSuccess(ctx *gin.Context, msg string, result any) {
	ctx.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: msg, Result: result})
}

// SendError
func SendError(ctx *gin.Context, msg string) {
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeError, Msg: msg})
}

// SendErrorf
func SendErrorf(ctx *gin.Context, format string, args ...any) {
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeError, Msg: fmt.Sprintf(format, args...)})
}
