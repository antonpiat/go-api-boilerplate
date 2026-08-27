package middleware

import (
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/authctx"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const loggerKey = "logger"

func Logger(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := GetRequestID(c)
		reqLogger := base.With(zap.String("request_id", reqID))
		c.Set(loggerKey, reqLogger)

		c.Next()

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
		}
		if p, ok := authctx.Get(c); ok {
			fields = append(fields, zap.String("user_id", p.UserID.String()))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}
		reqLogger.Info("http request", fields...)
	}
}

func LoggerFromContext(c *gin.Context, fallback *zap.Logger) *zap.Logger {
	if v, exists := c.Get(loggerKey); exists {
		if log, ok := v.(*zap.Logger); ok {
			return log
		}
	}
	return fallback
}
