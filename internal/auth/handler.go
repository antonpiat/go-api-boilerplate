package auth

import (
	"github.com/antonpiat/go-api-boilerplate/internal/authctx"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublic(rg *gin.RouterGroup) {
	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.Refresh)
}

func (h *Handler) RegisterProtected(rg *gin.RouterGroup) {
	rg.POST("/logout", h.Logout)
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	pair, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, pair)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	pair, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, pair)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, pair)
}

func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	p := authctx.MustGet(c)
	if err := h.svc.Logout(c.Request.Context(), p.JTI, req.RefreshToken, h.svc.AccessTTL()); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.NoContent(c)
}
