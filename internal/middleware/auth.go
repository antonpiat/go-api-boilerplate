package middleware

import (
	"context"
	"strings"

	"github.com/antonpiat/go-api-boilerplate/internal/authctx"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/gin-gonic/gin"
)

type AccessValidator interface {
	ValidateAccess(ctx context.Context, token string) (authctx.Principal, error)
}

const bearerPrefix = "Bearer "

func JWT(validator AccessValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			httpx.Fail(c, httpx.Unauthorized("missing authorization header"))
			c.Abort()
			return
		}
		if len(header) <= len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
			httpx.Fail(c, httpx.Unauthorized("invalid authorization header"))
			c.Abort()
			return
		}

		principal, err := validator.ValidateAccess(c.Request.Context(), header[len(bearerPrefix):])
		if err != nil {
			httpx.Fail(c, err)
			c.Abort()
			return
		}
		authctx.Set(c, principal)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		p, ok := authctx.Get(c)
		if !ok {
			httpx.Fail(c, httpx.Unauthorized("authentication required"))
			c.Abort()
			return
		}
		if _, ok := allowed[p.Role]; !ok {
			httpx.Fail(c, httpx.Forbidden("insufficient permissions"))
			c.Abort()
			return
		}
		c.Next()
	}
}
