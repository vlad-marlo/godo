package production

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// checkUserCredentials check auth data for user. Method return pointer to user and error.
// If where is user with email and password this method will return no error and current user object.
// Any else, service will return field error with correct data.
func (s *Service) checkUserCredentials(ctx context.Context, email, password string) (*model.User, error) {
	u, err := s.store.User().GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("get user by name", zap.Error(err))

		if errors.Is(err, store.ErrNotFound) {
			return nil, service.ErrBadAuthData.With(zap.Error(err))
		}

		return nil, service.ErrInternal.With(zap.Error(err))
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Pass), []byte(s.cfg.Server.Salt+password)) != nil {
		return nil, service.ErrBadAuthData
	}
	return u, nil
}

// CreateToken ...
func (s *Service) CreateToken(ctx context.Context, email, password, token string) (*model.CreateTokenResponse, error) {
	u, err := s.checkUserCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(token) {
	case BearerToken, JWTToken:
		return s.createJWTToken(u)
	case AuthToken, AuthorizationToken:
		return s.createAuthToken(ctx, u)
	}
	return nil, service.ErrBadTokenType.With(zap.String("token_type", token))
}

// jwtKeyFunc is helper func to get token encrypt key.
func (s *Service) jwtKeyFunc(*jwt.Token) (interface{}, error) {
	return []byte(s.cfg.Server.SecretKey), nil
}

// GetUserFromToken parses jwt token from raw token string.
func (s *Service) GetUserFromToken(ctx context.Context, t string) (uuid.UUID, error) {
	switch {
	case strings.HasPrefix(t, "Bearer "):
		return s.getUserFromJWTToken(ctx, t)
	case strings.HasPrefix(t, "Authorization "):
		return s.getUserFromAuthToken(ctx, t)
	default:
		return s.getUserFromAuthToken(ctx, t)
	}
}

// getUserFromAuthToken ...
func (s *Service) getUserFromAuthToken(ctx context.Context, t string) (uuid.UUID, error) {
	t = strings.TrimPrefix(strings.TrimPrefix(t, "Authorization "), "authorization ")
	u, err := s.store.Token().Get(ctx, t)

	if err != nil {
		if errors.Is(err, store.ErrUnknown) {
			return uuid.Nil, service.ErrInternal.With(zap.Error(err))
		}
		return uuid.Nil, service.ErrTokenNotValid.With(zap.Error(err))
	}

	if time.Now().UTC().After(u.ExpiresAt.UTC()) && u.Expires {
		return uuid.Nil, service.ErrTokenNotValid
	}

	return u.UserID, nil
}

// getUserFromJWTToken ...
func (s *Service) getUserFromJWTToken(ctx context.Context, t string) (uuid.UUID, error) {
	t = strings.TrimPrefix(t, "Bearer ")
	token, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, s.jwtKeyFunc)
	if err != nil {
		return uuid.Nil, service.ErrTokenNotValid.With(zap.Error(fmt.Errorf("parse jwt: %w", err)))
	}

	if !token.Valid {
		return uuid.Nil, service.ErrTokenNotValid
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, service.ErrTokenNotValid
	}

	var u uuid.UUID
	u, err = uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, service.ErrTokenNotValid.With(zap.Error(err))
	}

	if !s.store.User().Exists(ctx, u.String()) {
		return uuid.Nil, service.ErrTokenNotValid
	}

	return u, nil
}

// createJWTToken creates jwt token for user.
func (s *Service) createJWTToken(u *model.User) (*model.CreateTokenResponse, error) {
	t := time.Now()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   u.ID.String(),
		Audience:  []string{"access_token"},
		ExpiresAt: jwt.NewNumericDate(t.Add(s.cfg.Auth.AccessTokenLifeTime)),
		NotBefore: jwt.NewNumericDate(t),
		IssuedAt:  jwt.NewNumericDate(t),
		ID:        uuid.NewString(),
	})

	token, err := at.SignedString([]byte(s.cfg.Server.SecretKey))
	if err != nil {
		return nil, service.ErrInternal.With(zap.Error(fmt.Errorf("create jwt token: signed string: %w", err)))
	}

	return &model.CreateTokenResponse{
		TokenType:   BearerToken,
		AccessToken: token,
	}, nil
}

// createAuthToken create authorization token for user.
func (s *Service) createAuthToken(ctx context.Context, u *model.User) (*model.CreateTokenResponse, error) {
	now := time.Now()
	token, err := generateRandom(s.cfg.Auth.AuthTokenSize)
	if err != nil {
		return nil, service.ErrInternal.With(zap.Error(err), zap.String("summary", "error while generating auth token"))
	}
	t := &model.Token{
		UserID:    u.ID,
		Token:     token,
		ExpiresAt: now.Add(s.cfg.Auth.AccessTokenLifeTime),
		Expires:   false,
	}
	if err = s.store.Token().Create(ctx, t); err != nil {
		return nil, service.ErrInternal.With(zap.Error(err))
	}
	return &model.CreateTokenResponse{
		TokenType:   AuthorizationToken,
		AccessToken: t.Token,
	}, nil
}
