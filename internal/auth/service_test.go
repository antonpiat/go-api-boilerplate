package auth

import (
	"context"
	"testing"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/antonpiat/go-api-boilerplate/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type memUserRepo struct {
	byID    map[uuid.UUID]*user.User
	byEmail map[string]*user.User
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		byID:    make(map[uuid.UUID]*user.User),
		byEmail: make(map[string]*user.User),
	}
}

func (m *memUserRepo) Create(_ context.Context, u *user.User) error {
	if _, ok := m.byEmail[u.Email]; ok {
		return user.ErrEmailTaken
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}

func (m *memUserRepo) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*user.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) List(_ context.Context, _ httpx.Pagination) ([]user.User, int64, error) {
	out := make([]user.User, 0, len(m.byID))
	for _, u := range m.byID {
		out = append(out, *u)
	}
	return out, int64(len(out)), nil
}

func (m *memUserRepo) Update(_ context.Context, u *user.User) error {
	if _, ok := m.byID[u.ID]; !ok {
		return user.ErrNotFound
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}

func (m *memUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	u, ok := m.byID[id]
	if !ok {
		return user.ErrNotFound
	}
	delete(m.byID, id)
	delete(m.byEmail, u.Email)
	return nil
}

type memTokenStore struct {
	refresh   map[string]string
	blacklist map[string]struct{}
}

func newMemTokenStore() *memTokenStore {
	return &memTokenStore{
		refresh:   make(map[string]string),
		blacklist: make(map[string]struct{}),
	}
}

func (m *memTokenStore) SaveRefresh(_ context.Context, jti, userID string, _ time.Duration) error {
	m.refresh[jti] = userID
	return nil
}

func (m *memTokenStore) GetRefresh(_ context.Context, jti string) (string, error) {
	v, ok := m.refresh[jti]
	if !ok {
		return "", errNotFound("refresh")
	}
	return v, nil
}

func (m *memTokenStore) RevokeRefresh(_ context.Context, jti string) error {
	delete(m.refresh, jti)
	return nil
}

func (m *memTokenStore) BlacklistAccess(_ context.Context, jti string, _ time.Duration) error {
	m.blacklist[jti] = struct{}{}
	return nil
}

func (m *memTokenStore) IsAccessBlacklisted(_ context.Context, jti string) (bool, error) {
	_, ok := m.blacklist[jti]
	return ok, nil
}

type staticError string

func (e staticError) Error() string { return string(e) }

func errNotFound(kind string) error { return staticError(kind + " not found") }

func testService(t *testing.T) (*Service, *memUserRepo, *memTokenStore) {
	t.Helper()
	repo := newMemUserRepo()
	store := newMemTokenStore()
	tokens := NewTokenService(config.JWTConfig{
		AccessSecret:  "access-secret-for-tests-32bytes!",
		RefreshSecret: "refresh-secret-for-tests-32byte!",
		AccessTTL:     "15m",
		RefreshTTL:    "168h",
	}, "test")
	return NewService(repo, tokens, store), repo, store
}

func TestRegisterAndLogin(t *testing.T) {
	svc, _, _ := testService(t)
	ctx := context.Background()

	pair, err := svc.Register(ctx, RegisterRequest{Email: "Ada@Example.com", Password: "password123"})
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.Equal(t, "Bearer", pair.TokenType)

	_, err = svc.Register(ctx, RegisterRequest{Email: "ada@example.com", Password: "password123"})
	require.Error(t, err)
	appErr, ok := httpx.AsAppError(err)
	require.True(t, ok)
	require.Equal(t, httpx.CodeConflict, appErr.Code)

	login, err := svc.Login(ctx, LoginRequest{Email: "ada@example.com", Password: "password123"})
	require.NoError(t, err)
	require.NotEmpty(t, login.AccessToken)

	_, err = svc.Login(ctx, LoginRequest{Email: "ada@example.com", Password: "wrong-password"})
	require.Error(t, err)
	appErr, ok = httpx.AsAppError(err)
	require.True(t, ok)
	require.Equal(t, httpx.CodeUnauthorized, appErr.Code)
}

func TestRefreshRotatesAndLogoutBlacklists(t *testing.T) {
	svc, _, _ := testService(t)
	ctx := context.Background()

	pair, err := svc.Register(ctx, RegisterRequest{Email: "user@example.com", Password: "password123"})
	require.NoError(t, err)

	principal, err := svc.ValidateAccess(ctx, pair.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", principal.Email)

	rotated, err := svc.Refresh(ctx, pair.RefreshToken)
	require.NoError(t, err)
	require.NotEqual(t, pair.RefreshToken, rotated.RefreshToken)

	_, err = svc.Refresh(ctx, pair.RefreshToken)
	require.Error(t, err)

	err = svc.Logout(ctx, principal.JTI, rotated.RefreshToken, time.Minute)
	require.NoError(t, err)

	_, err = svc.ValidateAccess(ctx, pair.AccessToken)
	require.Error(t, err)
}
