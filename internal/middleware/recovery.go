package middleware

import (
	"net/http"

	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				LoggerFromContext(c, log).Error("panic recovered", zap.Any("panic", rec), zap.Stack("stack"))
				httpx.Fail(c, httpx.Internal("internal server error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}

func MaxBodySize(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if n > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		}
		c.Next()
	}
}
