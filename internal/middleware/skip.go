package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func skipInstrumentation(c *gin.Context) bool {
	if c.Request.Method == http.MethodOptions {
		return true
	}
	path := c.Request.URL.Path
	switch path {
	case "/metrics", "/health/live", "/health/ready", "/favicon.ico":
		return true
	}
	return strings.HasPrefix(path, "/swagger")
}
