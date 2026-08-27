package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/auth"
	"github.com/antonpiat/go-api-boilerplate/internal/cache"
	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/antonpiat/go-api-boilerplate/internal/database"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/antonpiat/go-api-boilerplate/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthAndUserFlow(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("set INTEGRATION=1 to run integration tests")
	}

	gin.SetMode(gin.TestMode)
	cfg := integrationConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.New(ctx, cfg.Database)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close(db) })

	rdb, err := cache.New(ctx, cfg.Redis)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rdb.Close() })

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	tokenSvc := auth.NewTokenService(cfg.JWT, cfg.App.Name)
	authSvc := auth.NewService(userRepo, tokenSvc, auth.NewRedisTokenStore(rdb))

	engine := New(Dependencies{
		Config: cfg,
		Logger: zap.NewNop(),
		Ready: CompositeReady{Checks: []func(context.Context) error{
			func(ctx context.Context) error { return database.Ping(ctx, db) },
			func(ctx context.Context) error { return cache.Ping(ctx, rdb) },
		}},
		AuthHandler: auth.NewHandler(authSvc),
		AuthService: authSvc,
		UserHandler: user.NewHandler(userSvc),
	})

	email := fmt.Sprintf("int-%s@example.com", uuid.NewString())
	adminEmail := fmt.Sprintf("admin-%s@example.com", uuid.NewString())

	reg := doJSON(t, engine, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusCreated, reg.Code)
	userTokens := decodeTokens(t, reg)

	me := doJSON(t, engine, http.MethodGet, "/api/v1/users/me", userTokens.AccessToken, nil)
	require.Equal(t, http.StatusOK, me.Code)

	listForbidden := doJSON(t, engine, http.MethodGet, "/api/v1/users", userTokens.AccessToken, nil)
	require.Equal(t, http.StatusForbidden, listForbidden.Code)

	adminReg := doJSON(t, engine, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email":    adminEmail,
		"password": "password123",
	})
	require.Equal(t, http.StatusCreated, adminReg.Code)
	require.NoError(t, db.Model(&user.User{}).Where("email = ?", adminEmail).Update("role", user.RoleAdmin).Error)

	adminLogin := doJSON(t, engine, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email":    adminEmail,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, adminLogin.Code)
	adminTokens := decodeTokens(t, adminLogin)

	listOK := doJSON(t, engine, http.MethodGet, "/api/v1/users?page=1&per_page=20", adminTokens.AccessToken, nil)
	require.Equal(t, http.StatusOK, listOK.Code)

	refresh := doJSON(t, engine, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": userTokens.RefreshToken,
	})
	require.Equal(t, http.StatusOK, refresh.Code)
	rotated := decodeTokens(t, refresh)

	logout := doJSON(t, engine, http.MethodPost, "/api/v1/auth/logout", rotated.AccessToken, map[string]string{
		"refresh_token": rotated.RefreshToken,
	})
	require.Equal(t, http.StatusNoContent, logout.Code)

	meAfterLogout := doJSON(t, engine, http.MethodGet, "/api/v1/users/me", rotated.AccessToken, nil)
	require.Equal(t, http.StatusUnauthorized, meAfterLogout.Code)
}

func integrationConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Metrics.Enabled = false
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	return cfg
}

func doJSON(t *testing.T, engine http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func decodeTokens(t *testing.T, rec *httptest.ResponseRecorder) auth.TokenPair {
	t.Helper()
	var env httpx.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	raw, err := json.Marshal(env.Data)
	require.NoError(t, err)
	var pair auth.TokenPair
	require.NoError(t, json.Unmarshal(raw, &pair))
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
	return pair
}
