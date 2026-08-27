package auth

import (
	"context"
	"errors"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/authctx"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/antonpiat/go-api-boilerplate/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users  user.Repository
	tokens *TokenService
	store  TokenStore
}

func NewService(users user.Repository, tokens *TokenService, store TokenStore) *Service {
	return &Service{users: users, tokens: tokens, store: store}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*TokenPair, error) {
	email := user.NormalizeEmail(req.Email)
	hash, err := user.HashPassword(req.Password)
	if err != nil {
		return nil, httpx.Internal("failed to hash password", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	u := &user.User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		Role:         user.RoleUser,
	}
	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			return nil, httpx.Conflict("email already registered")
		}
		return nil, httpx.Internal("failed to create user", err)
	}
	return s.issuePair(ctx, u)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, user.NormalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, httpx.Unauthorized("invalid email or password")
		}
		return nil, httpx.Internal("failed to lookup user", err)
	}
	if err := user.CheckPassword(u.PasswordHash, req.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, httpx.Unauthorized("invalid email or password")
		}
		return nil, httpx.Internal("failed to verify password", err)
	}
	return s.issuePair(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return nil, httpx.Unauthorized("invalid refresh token")
	}
	storedUserID, err := s.store.GetRefresh(ctx, claims.JTI)
	if err != nil || storedUserID != claims.UserID.String() {
		return nil, httpx.Unauthorized("invalid refresh token")
	}
	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, httpx.Unauthorized("invalid refresh token")
		}
		return nil, httpx.Internal("failed to lookup user", err)
	}
	_ = s.store.RevokeRefresh(ctx, claims.JTI)
	return s.issuePair(ctx, u)
}

func (s *Service) Logout(ctx context.Context, accessJTI, refreshToken string, accessTTL time.Duration) error {
	if refreshToken != "" {
		if claims, err := s.tokens.ParseRefresh(refreshToken); err == nil {
			_ = s.store.RevokeRefresh(ctx, claims.JTI)
		}
	}
	if accessJTI != "" {
		if err := s.store.BlacklistAccess(ctx, accessJTI, accessTTL); err != nil {
			return httpx.Internal("failed to revoke access token", err)
		}
	}
	return nil
}

func (s *Service) ValidateAccess(ctx context.Context, token string) (authctx.Principal, error) {
	claims, err := s.tokens.ParseAccess(token)
	if err != nil {
		return authctx.Principal{}, httpx.Unauthorized("invalid or expired access token")
	}
	blacklisted, err := s.store.IsAccessBlacklisted(ctx, claims.JTI)
	if err != nil {
		return authctx.Principal{}, httpx.Internal("failed to validate token", err)
	}
	if blacklisted {
		return authctx.Principal{}, httpx.Unauthorized("token has been revoked")
	}
	return authctx.Principal{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		JTI:    claims.JTI,
	}, nil
}

func (s *Service) issuePair(ctx context.Context, u *user.User) (*TokenPair, error) {
	access, _, err := s.tokens.IssueAccess(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, httpx.Internal("failed to issue access token", err)
	}
	refresh, refreshJTI, err := s.tokens.IssueRefresh(u.ID)
	if err != nil {
		return nil, httpx.Internal("failed to issue refresh token", err)
	}
	if err := s.store.SaveRefresh(ctx, refreshJTI, u.ID.String(), s.tokens.RefreshTTL()); err != nil {
		return nil, httpx.Internal("failed to persist refresh token", err)
	}
	return newTokenPair(access, refresh, s.tokens.AccessTTL()), nil
}

func (s *Service) AccessTTL() time.Duration {
	return s.tokens.AccessTTL()
}
