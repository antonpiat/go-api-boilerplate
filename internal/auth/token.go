package auth

import (
	"errors"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

var (
	errInvalidToken     = errors.New("invalid token")
	errInvalidTokenType = errors.New("invalid token type")
	errInvalidSubject   = errors.New("invalid subject")
	errMissingJTI       = errors.New("missing jti")
)

type jwtClaims struct {
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
	Typ   string `json:"typ"`
	jwt.RegisteredClaims
}

type TokenService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	parser        *jwt.Parser
	accessKey     jwt.Keyfunc
	refreshKey    jwt.Keyfunc
}

func NewTokenService(cfg config.JWTConfig, issuer string) *TokenService {
	s := &TokenService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTLDuration(),
		refreshTTL:    cfg.RefreshTTLDuration(),
		issuer:        issuer,
		parser:        jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})),
	}
	s.accessKey = func(*jwt.Token) (any, error) { return s.accessSecret, nil }
	s.refreshKey = func(*jwt.Token) (any, error) { return s.refreshSecret, nil }
	return s
}

func (s *TokenService) AccessTTL() time.Duration  { return s.accessTTL }
func (s *TokenService) RefreshTTL() time.Duration { return s.refreshTTL }

func (s *TokenService) IssueAccess(userID uuid.UUID, email, role string) (token string, jti string, err error) {
	return s.issue(s.accessSecret, tokenTypeAccess, userID, email, role, s.accessTTL)
}

func (s *TokenService) IssueRefresh(userID uuid.UUID) (token string, jti string, err error) {
	return s.issue(s.refreshSecret, tokenTypeRefresh, userID, "", "", s.refreshTTL)
}

func (s *TokenService) ParseAccess(token string) (*TokenClaims, error) {
	return s.parse(token, s.accessKey, tokenTypeAccess)
}

func (s *TokenService) ParseRefresh(token string) (*TokenClaims, error) {
	return s.parse(token, s.refreshKey, tokenTypeRefresh)
}

func (s *TokenService) issue(secret []byte, typ string, userID uuid.UUID, email, role string, ttl time.Duration) (string, string, error) {
	jti := uuid.NewString()
	now := time.Now().UTC()
	claims := jwtClaims{
		Email: email,
		Role:  role,
		Typ:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(secret)
	if err != nil {
		return "", "", err
	}
	return signed, jti, nil
}

func (s *TokenService) parse(tokenString string, key jwt.Keyfunc, expectedType string) (*TokenClaims, error) {
	token, err := s.parser.ParseWithClaims(tokenString, &jwtClaims{}, key)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, errInvalidToken
	}
	if claims.Typ != expectedType {
		return nil, errInvalidTokenType
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, errInvalidSubject
	}
	if claims.ID == "" {
		return nil, errMissingJTI
	}
	return &TokenClaims{
		UserID: userID,
		Email:  claims.Email,
		Role:   claims.Role,
		JTI:    claims.ID,
	}, nil
}
