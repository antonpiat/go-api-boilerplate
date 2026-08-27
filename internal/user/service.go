package user

import (
	"context"
	"errors"
	"strings"

	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return u, nil
}

func (s *Service) List(ctx context.Context, page httpx.Pagination) ([]PublicUser, int64, error) {
	users, total, err := s.repo.List(ctx, page)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list users", err)
	}
	out := make([]PublicUser, 0, len(users))
	for i := range users {
		out = append(out, users[i].ToPublic())
	}
	return out, total, nil
}

func (s *Service) UpdateMe(ctx context.Context, id uuid.UUID, req UpdateMeRequest) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if err := applyProfileUpdate(u, req.Email, req.Password); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, mapRepoError(err)
	}
	return u, nil
}

func (s *Service) AdminUpdate(ctx context.Context, actorID, targetID uuid.UUID, req AdminUpdateRequest) (*User, error) {
	u, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if err := applyProfileUpdate(u, req.Email, req.Password); err != nil {
		return nil, err
	}
	if req.Role != nil {
		if actorID == targetID && *req.Role != RoleAdmin {
			return nil, httpx.Forbidden("cannot demote your own admin role")
		}
		u.Role = *req.Role
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, mapRepoError(err)
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, actorID, targetID uuid.UUID) error {
	if actorID == targetID {
		return httpx.Forbidden("cannot delete your own account")
	}
	if err := s.repo.Delete(ctx, targetID); err != nil {
		return mapRepoError(err)
	}
	return nil
}

func applyProfileUpdate(u *User, email *string, password *string) error {
	if email != nil {
		u.Email = NormalizeEmail(*email)
	}
	if password != nil {
		hash, err := HashPassword(*password)
		if err != nil {
			return httpx.Internal("failed to hash password", err)
		}
		u.PasswordHash = hash
	}
	return nil
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func mapRepoError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return httpx.NotFound("user not found")
	case errors.Is(err, ErrEmailTaken):
		return httpx.Conflict("email already registered")
	default:
		return httpx.Internal("user operation failed", err)
	}
}
