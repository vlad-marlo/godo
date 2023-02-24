package production

// TODO: add more const for error messages.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"strings"
	"time"
)

const (
	BearerToken        = "bearer"
	JWTToken           = "jwt"
	AuthorizationToken = "authorization"
	AuthToken          = "auth"
)

// RegisterUser ...
func (s *Service) RegisterUser(ctx context.Context, email, password string) (*model.User, error) {
	// validate password.
	if err := passwordvalidator.Validate(password, s.cfg.Auth.PasswordDifficult); err != nil {
		return nil, service.ErrPasswordToEasy.With(zap.Error(err))
	}

	// check email.
	ea, err := mail.ParseAddress(email)
	if err != nil {
		return nil, service.ErrEmailNotValid.With(zap.Error(err))
	}
	email = ea.Address

	// take hash of password.
	// TODO: fix bug with too long password (split password to parts by length and encrypt them separately).
	var pass []byte
	pass, err = bcrypt.GenerateFromPassword(
		[]byte(s.cfg.Server.Salt+password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, service.ErrPasswordToLong.With(zap.Error(err), zap.String("email", email))
	}

	u := &model.User{
		ID:    uuid.New(),
		Pass:  string(pass),
		Email: email,
	}
	if err = s.store.User().Create(ctx, u); err != nil {
		if errors.Is(err, store.ErrUserAlreadyExists) {
			return nil, service.ErrEmailAlreadyInUse
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}
	u.Pass = ""

	return u, nil
}

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
func (s *Service) GetUserFromToken(ctx context.Context, t string) (string, error) {
	if strings.HasPrefix(t, "Bearer ") {
		return s.getUserFromJWTToken(ctx, t)
	}
	if strings.HasPrefix(t, "Authorization ") {
		return s.getUserFromAuthToken(ctx, t)
	}
	return "", service.ErrBadAuthData
}

// getUserFromAuthToken ...
func (s *Service) getUserFromAuthToken(ctx context.Context, t string) (string, error) {
	t = strings.TrimPrefix(t, "Authorization ")
	u, err := s.store.Token().Get(ctx, t)
	if err != nil {
		return "", service.ErrTokenNotValid.With(zap.Error(err))
	}
	return u.UserID.String(), nil
}

// getUserFromJWTToken ...
func (s *Service) getUserFromJWTToken(ctx context.Context, t string) (string, error) {
	t = strings.TrimPrefix(t, "Bearer ")
	token, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, s.jwtKeyFunc)
	if err != nil {
		return "", service.ErrInternal.With(zap.Error(fmt.Errorf("parse jwt: %w", err)))
	}

	if !token.Valid {
		return "", service.ErrTokenNotValid
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", service.ErrTokenNotValid
	}

	u := claims.Subject
	if !s.store.User().Exists(ctx, u) {
		return "", service.ErrTokenNotValid
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
	if err := s.store.Token().Create(ctx, t); err != nil {
		return nil, service.ErrInternal.With(zap.Error(err))
	}
	return &model.CreateTokenResponse{
		TokenType:   AuthorizationToken,
		AccessToken: t.Token,
	}, nil
}

// generateRandom generates new string with provided size.
func generateRandom(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
