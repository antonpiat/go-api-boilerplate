package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/auth"
	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/antonpiat/go-api-boilerplate/internal/server"
	"github.com/antonpiat/go-api-boilerplate/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := config.DefaultConfig()
	cfg.Metrics.Enabled = false
	repo := rejectRepo{}
	tokens := auth.NewTokenService(cfg.JWT, "test")
	svc := auth.NewService(repo, tokens, noopStore{})
	return server.New(server.Dependencies{
		Config:      cfg,
		Logger:      zap.NewNop(),
		AuthHandler: auth.NewHandler(svc),
		AuthService: svc,
		UserHandler: user.NewHandler(user.NewService(repo)),
	})
}

func TestRegisterValidation(t *testing.T) {
	engine := testEngine(t)
	body, err := json.Marshal(map[string]string{"email": "not-an-email", "password": "short"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHealthLive(t *testing.T) {
	engine := testEngine(t)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"success":true`)
}

func TestCompositeReadyReportsFirstError(t *testing.T) {
	ready := server.CompositeReady{Checks: []func(context.Context) error{
		func(context.Context) error { return nil },
		func(context.Context) error { return errors.New("redis down") },
	}}
	require.EqualError(t, ready.Ready(context.Background()), "redis down")

	ok := server.CompositeReady{Checks: []func(context.Context) error{
		func(context.Context) error { return nil },
		func(context.Context) error { return nil },
	}}
	require.NoError(t, ok.Ready(context.Background()))
}

func TestRootRedirectsToSwagger(t *testing.T) {
	engine := testEngine(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/swagger/index.html", rec.Header().Get("Location"))
}

func TestUnknownRouteJSON(t *testing.T) {
	engine := testEngine(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "NOT_FOUND")
}

func TestUsersMeRequiresAuth(t *testing.T) {
	engine := testEngine(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

type rejectRepo struct{}

func (rejectRepo) Create(_ context.Context, _ *user.User) error { return user.ErrEmailTaken }
func (rejectRepo) GetByID(_ context.Context, _ uuid.UUID) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (rejectRepo) GetByEmail(_ context.Context, _ string) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (rejectRepo) List(_ context.Context, _ httpx.Pagination) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (rejectRepo) Update(_ context.Context, _ *user.User) error { return nil }
func (rejectRepo) Delete(_ context.Context, _ uuid.UUID) error  { return user.ErrNotFound }

type noopStore struct{}

func (noopStore) SaveRefresh(_ context.Context, _, _ string, _ time.Duration) error {
	return nil
}
func (noopStore) GetRefresh(_ context.Context, _ string) (string, error) {
	return "", errors.New("not found")
}
func (noopStore) RevokeRefresh(_ context.Context, _ string) error { return nil }
func (noopStore) BlacklistAccess(_ context.Context, _ string, _ time.Duration) error {
	return nil
}
func (noopStore) IsAccessBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
