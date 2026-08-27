package httpx

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/x?page=2&per_page=500", nil)

	p := ParsePagination(c)
	require.Equal(t, 2, p.Page)
	require.Equal(t, MaxPerPage, p.PerPage)
	require.Equal(t, MaxPerPage, p.Offset())
}
