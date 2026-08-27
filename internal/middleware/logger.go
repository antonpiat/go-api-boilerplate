package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const loggerKey = "logger"

func Logger(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipInstrumentation(c) {
			c.Next()
			return
		}

		start := time.Now()
		reqID := GetRequestID(c)
		reqLogger := base.With(zap.String("request_id", reqID))
		c.Set(loggerKey, reqLogger)

		c.Next()

		status := c.Writer.Status()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}
		msg := fmt.Sprintf("%s %s %d %dms",
			c.Request.Method,
			path,
			status,
			time.Since(start).Milliseconds(),
		)
		if len(reqID) >= 8 {
			msg += " " + reqID[:8]
		}

		switch {
		case status >= http.StatusInternalServerError:
			reqLogger.Error(msg)
		case status >= http.StatusBadRequest:
			reqLogger.Warn(msg)
		default:
			base.Info(msg)
		}
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
