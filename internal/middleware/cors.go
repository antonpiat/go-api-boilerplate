package middleware

import (
	"net/http"
	"strings"

	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/gin-gonic/gin"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowAll := len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*"
	methods := strings.Join(cfg.AllowMethods, ",")
	headers := strings.Join(cfg.AllowHeaders, ",")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if allowAll {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if originAllowed(origin, cfg.AllowOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", methods)
			c.Header("Access-Control-Allow-Headers", headers)
			if cfg.AllowCredentials && !allowAll {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, o := range allowed {
		if o == origin {
			return true
		}
	}
	return false
}
