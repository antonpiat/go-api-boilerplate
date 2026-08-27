package user

import (
	"context"
	"testing"

	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type memRepo struct {
	users map[uuid.UUID]*User
}

func newMemRepo(users ...*User) *memRepo {
	m := &memRepo{users: make(map[uuid.UUID]*User)}
	for _, u := range users {
		cp := *u
		m.users[u.ID] = &cp
	}
	return m
}

func (m *memRepo) Create(_ context.Context, u *User) error {
	if _, ok := m.users[u.ID]; ok {
		return ErrEmailTaken
	}
	for _, existing := range m.users {
		if existing.Email == u.Email {
			return ErrEmailTaken
		}
	}
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memRepo) GetByEmail(_ context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *memRepo) List(_ context.Context, _ httpx.Pagination) ([]User, int64, error) {
	out := make([]User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) Update(_ context.Context, u *User) error {
	if _, ok := m.users[u.ID]; !ok {
		return ErrNotFound
	}
	cp := *u
	m.users[u.ID] = &cp
	return nil
}

func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.users[id]; !ok {
		return ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func TestUpdateMeAndAdminGuards(t *testing.T) {
	adminID := uuid.New()
	userID := uuid.New()
	repo := newMemRepo(
		&User{ID: adminID, Email: "admin@example.com", Role: RoleAdmin, PasswordHash: "x"},
		&User{ID: userID, Email: "user@example.com", Role: RoleUser, PasswordHash: "x"},
	)
	svc := NewService(repo)
	ctx := context.Background()

	email := "user2@example.com"
	updated, err := svc.UpdateMe(ctx, userID, UpdateMeRequest{Email: &email})
	require.NoError(t, err)
	require.Equal(t, "user2@example.com", updated.Email)

	role := RoleUser
	_, err = svc.AdminUpdate(ctx, adminID, adminID, AdminUpdateRequest{Role: &role})
	require.Error(t, err)
	appErr, ok := httpx.AsAppError(err)
	require.True(t, ok)
	require.Equal(t, httpx.CodeForbidden, appErr.Code)

	err = svc.Delete(ctx, adminID, adminID)
	require.Error(t, err)

	err = svc.Delete(ctx, adminID, userID)
	require.NoError(t, err)
}

func TestListMapsPublicUsers(t *testing.T) {
	u := &User{ID: uuid.New(), Email: "a@example.com", Role: RoleUser, PasswordHash: "secret"}
	svc := NewService(newMemRepo(u))
	out, total, err := svc.List(context.Background(), httpx.Pagination{Page: 1, PerPage: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, out, 1)
	require.Equal(t, "a@example.com", out[0].Email)
}
