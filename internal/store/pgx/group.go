package pgx

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

var _ store.GroupRepository = (*GroupRepository)(nil)

// GroupRepository ...
type GroupRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewGroupRepository return new instance of GroupRepository.
func NewGroupRepository(cli Client) *GroupRepository {
	return &GroupRepository{
		pool: cli.P(),
		log:  cli.L(),
	}
}

// Create created new record about group.
//
// Errors:
// store.ErrGroupAlreadyExists	group with provided ID/Name already exists;
// store.ErrBadData bad foreign key to create group;
func (repo *GroupRepository) Create(ctx context.Context, group *model.Group) error {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		repo.log.Error("failed to begin transaction: check pgx driver", TraceError(err)...)
		return store.ErrUnknown
	}

	defer func() {
		if err = tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			repo.log.Error("failed to rollback transaction: check pgx driver", TraceError(err)...)
		}
	}()

	if err = tx.QueryRow(
		ctx,
		`INSERT INTO groups(id, "name", description, "owner") VALUES ($1, $2, $3, $4) RETURNING created_at;`,
		group.ID,
		group.Name,
		group.Description,
		group.Owner,
	).Scan(&group.CreatedAt); err != nil {

		repo.log.Log(_unknownLevel, "create user", TraceError(err)...)

		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrUniqueViolation
			case pgerrcode.InvalidForeignKey, pgerrcode.ForeignKeyViolation:
				return store.ErrFKViolation
			}

		}

		return Unknown(err)
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO roles(members, tasks, reviews, "comments")  VALUES (0, 0, 0, 0) ON CONFLICT DO NOTHING;`,
	); err != nil {
		repo.log.Warn("unknown error while creating role", TraceError(err)...)
		return Unknown(err)
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO user_in_group(user_id, group_id, role_id, is_admin)
VALUES ($1, $2, (SELECT r.id FROM roles r WHERE r.members = 0 AND r.comments = 0 AND r.reviews = 0 AND r.tasks = 0),
        TRUE);`,
		group.Owner,
		group.ID,
	); err != nil {
		repo.log.Error("unknown error while creating record about user in group", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	if err = tx.Commit(ctx); err != nil {
		repo.log.Error("failed to commit transaction: check pgx driver", TraceError(err)...)
		return store.ErrUnknown
	}

	return nil
}

// UserExists ...
func (repo *GroupRepository) UserExists(ctx context.Context, group, user string) (ok bool) {
	if err := repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM user_in_group WHERE group_id = $1 AND user_id = $2);`,
		group,
		user,
	).Scan(&ok); err != nil {
		repo.log.Log(_unknownLevel, "get user existence in group", TraceError(err)...)
	}
	return
}

// GetRoleOfMember return role of user in provided group.
func (repo *GroupRepository) GetRoleOfMember(ctx context.Context, user, group uuid.UUID) (role *model.Role, err error) {
	role = new(model.Role)
	if err = repo.pool.QueryRow(
		ctx,
		`SELECT r.id,
       CASE WHEN uig.is_admin THEN 100 ELSE r.members END,
       CASE WHEN uig.is_admin THEN 100 ELSE r.tasks END,
       CASE WHEN uig.is_admin THEN 100 ELSE r.reviews END,
       CASE WHEN uig.is_admin THEN 100 ELSE r.comments END
FROM roles r
         JOIN user_in_group uig on r.id = uig.role_id
WHERE uig.user_id = $1
  and uig.group_id = $2;`,
		user,
		group,
	).Scan(&role.ID, &role.Members, &role.Tasks, &role.Reviews, &role.Comments); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		repo.log.Log(_unknownLevel, "get role of user in group", TraceError(err)...)
		return nil, Unknown(err)
	}
	return
}

// GetByUser ...
func (repo *GroupRepository) GetByUser(ctx context.Context, user uuid.UUID) (groups []*model.Group, err error) {
	q := `SELECT g.id, g.name, g.description, g.owner, g.created_at
FROM groups g
         JOIN user_in_group uig on g.id = uig.group_id
WHERE uig.user_id = $1;`
	var rows pgx.Rows

	rows, err = repo.pool.Query(
		ctx,
		q,
		user,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		repo.log.Log(_unknownLevel, "get groups by user", TraceError(err)...)
		return nil, Unknown(err)
	}
	defer rows.Close()

	for rows.Next() {
		g := new(model.Group)

		if err = rows.Scan(&g.ID, &g.Name, &g.Description, &g.Owner, &g.CreatedAt); err != nil {
			repo.log.Warn("unknown error while scanning group while getting all groups by user", TraceError(err)...)
			return nil, Unknown(err)
		}

		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, Unknown(err)
	}
	return
}

// Get return group by id
func (repo *GroupRepository) Get(ctx context.Context, id uuid.UUID) (*model.Group, error) {
	g := new(model.Group)
	if err := repo.pool.QueryRow(
		ctx,
		`SELECT g.id, g.name, g.description, g.created_at, g.owner FROM groups g WHERE g.id = $1`,
		id,
	).Scan(
		&g.ID,
		&g.Name,
		&g.Description,
		&g.CreatedAt,
		&g.Owner,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		repo.log.Log(_unknownLevel, "get group by id", TraceError(err)...)
		return nil, Unknown(err)
	}
	return g, nil

}
