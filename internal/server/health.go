package server

import (
	"context"
	"net/http"

	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/gin-gonic/gin"
)

type ReadyChecker interface {
	Ready(ctx context.Context) error
}

type HealthHandler struct {
	ready ReadyChecker
}

func NewHealthHandler(ready ReadyChecker) *HealthHandler {
	return &HealthHandler{ready: ready}
}

func (h *HealthHandler) Live(c *gin.Context) {
	httpx.OK(c, gin.H{"status": "ok"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if h.ready == nil {
		httpx.OK(c, gin.H{"status": "ok"})
		return
	}
	if err := h.ready.Ready(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, httpx.Envelope{
			Success: false,
			Error: &httpx.ErrorBody{
				Code:    httpx.CodeInternal,
				Message: "not ready",
			},
		})
		return
	}
	httpx.OK(c, gin.H{"status": "ok"})
}

type CompositeReady struct {
	Checks []func(ctx context.Context) error
}

func (c CompositeReady) Ready(ctx context.Context) error {
	for _, check := range c.Checks {
		if err := check(ctx); err != nil {
			return err
		}
	}
	return nil
}
