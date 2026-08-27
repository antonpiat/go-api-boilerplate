package server

import (
	"context"
	"net/http"

	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/gin-gonic/gin"
)

const jsonContentType = "application/json; charset=utf-8"

var liveOKBody = []byte(`{"success":true,"data":{"status":"ok"}}`)

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
	c.Data(http.StatusOK, jsonContentType, liveOKBody)
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if h.ready == nil {
		c.Data(http.StatusOK, jsonContentType, liveOKBody)
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
	c.Data(http.StatusOK, jsonContentType, liveOKBody)
}

type CompositeReady struct {
	Checks []func(ctx context.Context) error
}

func (c CompositeReady) Ready(ctx context.Context) error {
	n := len(c.Checks)
	switch n {
	case 0:
		return nil
	case 1:
		return c.Checks[0](ctx)
	}

	errCh := make(chan error, n)
	for _, check := range c.Checks {
		go func(fn func(context.Context) error) {
			errCh <- fn(ctx)
		}(check)
	}

	var first error
	for i := 0; i < n; i++ {
		if err := <-errCh; err != nil && first == nil {
			first = err
		}
	}
	return first
}
