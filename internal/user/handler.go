package user

import (
	"net/http"

	"github.com/antonpiat/go-api-boilerplate/internal/authctx"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/me", h.Me)
	rg.PATCH("/me", h.UpdateMe)
}

func (h *Handler) RegisterAdmin(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.GET("/:id", h.Get)
	rg.PATCH("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}

func (h *Handler) Me(c *gin.Context) {
	p := authctx.MustGet(c)
	u, err := h.svc.GetByID(c.Request.Context(), p.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u.ToPublic())
}

func (h *Handler) UpdateMe(c *gin.Context) {
	var req UpdateMeRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	p := authctx.MustGet(c)
	u, err := h.svc.UpdateMe(c.Request.Context(), p.UserID, req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u.ToPublic())
}

func (h *Handler) List(c *gin.Context) {
	page := httpx.ParsePagination(c)
	users, total, err := h.svc.List(c.Request.Context(), page)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.JSONPage(c, http.StatusOK, users, httpx.NewPageMeta(page, total))
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u.ToPublic())
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req AdminUpdateRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	p := authctx.MustGet(c)
	u, err := h.svc.AdminUpdate(c.Request.Context(), p.UserID, id, req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, u.ToPublic())
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	p := authctx.MustGet(c)
	if err := h.svc.Delete(c.Request.Context(), p.UserID, id); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.NoContent(c)
}

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, httpx.Validation("invalid user id", httpx.FieldError{Field: "id", Message: "must be a valid UUID"})
	}
	return id, nil
}
