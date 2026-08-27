package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Meta    *PageMeta  `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    ErrorCode    `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Success: true, Data: data})
}

func JSONPage(c *gin.Context, status int, data any, meta PageMeta) {
	c.JSON(status, Envelope{Success: true, Data: data, Meta: &meta})
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	JSON(c, http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Fail(c *gin.Context, err error) {
	appErr, ok := AsAppError(err)
	if !ok {
		appErr = Internal("internal server error", err)
	}

	if appErr.Status >= http.StatusInternalServerError {
		if log, exists := c.Get("logger"); exists {
			if zapLog, ok := log.(*zap.Logger); ok {
				zapLog.Error("request failed", zap.Error(appErr.Unwrap()), zap.String("message", appErr.Message))
			}
		}
		c.JSON(appErr.Status, Envelope{
			Success: false,
			Error: &ErrorBody{
				Code:    appErr.Code,
				Message: "internal server error",
			},
		})
		return
	}

	c.JSON(appErr.Status, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Fields,
		},
	})
}
