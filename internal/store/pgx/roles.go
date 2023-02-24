package pgx

import (
	"context"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

type RoleRepository struct {
	p *pgxpool.Pool
	l *zap.Logger
}

// NewRoleRepository ...
func NewRoleRepository(cli Client) *RoleRepository {
	return &RoleRepository{cli.P(), cli.L()}
}

func (repo *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	q := `INSERT INTO roles
(create_task, read_task, update_task, delete_task, create_issue, read_issue,
 update_issue, review_task, read_members, invite_members, delete_members)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT DO NOTHING
RETURNING id;`

	err := repo.p.QueryRow(
		ctx,
		q,
		role.CreateTask, role.ReadTask, role.UpdateTask, role.DeleteTask, role.CreateIssue, role.ReadIssue, role.UpdateIssue, role.ReviewTask,
		role.ReadMembers, role.InviteMembers, role.DeleteMembers,
	).Scan(&role.ID)
	if err != nil {
		repo.l.Error("unknown error", TraceError(err)...)

		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidForeignKey, pgerrcode.ForeignKeyViolation:
				return store.ErrBadForeignKey
			}
		}

		return store.ErrUnknown
	}

	return nil
}

func (repo *RoleRepository) ChangeUserRole(ctx context.Context, role *model.UserInGroup) error {
	q := `INSERT INTO user_in_group(user_id, group_id, role_id)
VALUES ($1, $2, $3)
ON CONFLICT DO UPDATE
    SET role_id = $3
WHERE user_id = $1
  AND group_id = $2;`

	_, err := repo.p.Exec(ctx, q, role.User, role.Group, role.Role)
	if err != nil {
		repo.l.Error("unknown error while creating group_user relation", TraceError(err)...)
		return store.ErrUnknown
	}
	return nil
}
