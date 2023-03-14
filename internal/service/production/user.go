package production

// TODO: add more const for error messages.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/service"
	"github.com/vlad-marlo/godo/internal/store"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
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

// CreateInvite ...
func (s *Service) CreateInvite(ctx context.Context, user uuid.UUID, group uuid.UUID, role *model.Role, limit int) (*model.CreateInviteResponse, error) {
	if limit <= 0 {
		return nil, service.ErrBadInviteLimit
	}
	userRole, err := s.store.Group().GetRoleOfMember(ctx, user, group)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, service.ErrForbidden
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}

	if userRole.Members < 2 {
		return nil, service.ErrForbidden
	}

	invite := uuid.New()

	s.store.Group()

	if err = s.store.Role().Get(ctx, role); err != nil {
		return nil, service.ErrInternal.With(zap.Error(err))
	}

	if err = s.store.Invite().Create(ctx, invite, role.ID, group, limit); err != nil {

		switch errors.Unwrap(err) {

		case store.ErrUniqueViolation:
			return nil, service.ErrConflict

		case store.ErrFKViolation:
			return nil, service.ErrBadData

		default:
			return nil, service.ErrInternal.With(zap.Error(err))
		}
	}
	return &model.CreateInviteResponse{
		// invite link template
		Link:  fmt.Sprintf(s.cfg.Server.InviteLinkTemplate, s.cfg.Server.BaseURL, group, invite),
		Limit: limit,
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

// GetMe ...
func (s *Service) GetMe(ctx context.Context, user uuid.UUID) (*model.GetMeResponse, error) {
	u, err := s.store.User().Get(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, service.ErrUserNotFound
		case errors.Is(err, store.ErrUnknown):
			return nil, service.ErrInternal.With(zap.Error(err))
		default:
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}

	res := &model.GetMeResponse{
		ID:     user,
		Email:  u.Email,
		Groups: []model.GroupInUser{},
	}

	var groups []*model.Group
	groups, err = s.store.Group().GetByUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			return nil, service.ErrNotFound
		default:
		}
		return nil, service.ErrInternal.With(zap.Error(err))
	}

	for _, group := range groups {
		var tasks []*model.Task

		tasks, err = s.store.Task().AllByGroupAndUser(ctx, group.ID, user)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, service.ErrNotFound
			}
			return nil, service.ErrInternal.With(zap.Error(err))
		}

		res.Groups = append(res.Groups, model.GroupInUser{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Tasks:       tasks,
		})

	}

	return res, nil
}
