package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func (p Pagination) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

type PageMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

func ParsePagination(c *gin.Context) Pagination {
	page := parsePositiveInt(c.Query("page"), DefaultPage)
	perPage := parsePositiveInt(c.Query("per_page"), DefaultPerPage)
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return Pagination{Page: page, PerPage: perPage}
}

func NewPageMeta(p Pagination, total int64) PageMeta {
	return PageMeta{Page: p.Page, PerPage: p.PerPage, Total: total}
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
