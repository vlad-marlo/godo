package production

// TODO: add more const for error messages.

import (
	"context"
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
	"time"
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

// LoginUserJWT ...
func (s *Service) LoginUserJWT(ctx context.Context, email, password string) (*model.CreateJWTResponse, error) {
	u, err := s.store.User().GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("get user by name", zap.Error(err))
		if errors.Is(err, store.ErrNotFound) {
			return nil, service.ErrBadAuthData.With(zap.Error(err))
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Pass), []byte(s.cfg.Server.Salt+password)) != nil {
		return nil, service.ErrBadAuthData.With()
	}
	at, rt, err := s.createJWTToken(u)
	if err != nil {
		s.log.Error("create jwt token", zap.Error(err))
		return nil, service.ErrInternal.With(zap.Error(fmt.Errorf("store: create jwt token: %w", err)))
	}

	return &model.CreateJWTResponse{
		TokenType:    "bearer",
		AccessToken:  at,
		RefreshToken: rt,
	}, nil
}

// GetUserFromToken parses jwt token from raw token string.
func (s *Service) GetUserFromToken(ctx context.Context, t string) (string, error) {
	token, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Server.SecretKey), nil
	})
	if err != nil {
		return "", service.ErrInternal.With(zap.Error(fmt.Errorf("parse jwt: %w", err)))
	}
	if !token.Valid {
		return "", service.ErrTokenNotValid
	}
	rc, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", service.ErrTokenNotValid
	}
	u := rc.Subject
	if !s.store.User().Exists(ctx, u) {
		return "", service.ErrTokenNotValid
	}
	return u, nil
}

// createJWTToken creates access and refresh tokens for user.
// TODO: maybe split logic of creating access and refresh token.
func (s *Service) createJWTToken(u *model.User) (string, string, error) {
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   u.ID.String(),
		Audience:  []string{"access_token"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.Auth.AccessTokenLifeTime)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	})
	ats, err := at.SignedString([]byte(s.cfg.Server.SecretKey))
	if err != nil {
		return "", "", fmt.Errorf("error while creating access token: jwt: token: signed string: %w", err)
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   u.ID.String(),
		Audience:  []string{"refresh_token"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.Auth.RefreshTokenLifeTime)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	})
	rts, err := rt.SignedString([]byte(s.cfg.Server.SecretKey))
	if err != nil {
		s.log.Debug("error while getting signed string for refresh token", zap.Error(err))
		return "", "", fmt.Errorf("creating refresh token: jwt: token: signed string: %w", err)
	}
	return ats, rts, nil
}
