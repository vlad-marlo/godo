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

// NewGroupRepository return new instance of GroupRepository
func NewGroupRepository(cli Client) *GroupRepository {
	return &GroupRepository{
		pool: cli.P(),
		log:  cli.L(),
	}
}

// Exists ...
func (repo *GroupRepository) Exists(ctx context.Context, id string) (ok bool) {
	_ = repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM groups WHERE id = $1);`,
		id,
	).Scan(&ok)
	return ok
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
		`INSERT INTO groups(id, name, description, owner) VALUES ($1, $2, $3, $4) RETURNING created_at;`,
		group.ID,
		group.Name,
		group.Description,
		group.Owner,
	).Scan(&group.CreatedAt); err != nil {

		repo.log.Debug(
			"create user: pgx error",
			TraceError(err)...,
		)

		if pgErr, ok := err.(*pgconn.PgError); ok {

			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrGroupAlreadyExists
			}

			if pgErr.Code == pgerrcode.InvalidForeignKey || pgerrcode.ForeignKeyViolation == pgErr.Code {
				return store.ErrBadData
			}
		}

		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO roles(members, tasks, reviews, "comments")  VALUES (0, 0, 0, 0) ON CONFLICT DO NOTHING;`,
	); err != nil {
		repo.log.Error("unknown error while creating role", TraceError(err)...)
		return store.ErrUnknown
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO user_in_group(user_id, group_id, role_id, is_admin)
VALUES ($1, $2, (SELECT x.id
                 FROM roles x
                 WHERE x.comments = 0
                   AND x.reviews = 0
                   AND x.members = 0
                   AND x.tasks = 0
                 LIMIT 1), TRUE);`,
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

// Get return group with provided id.
// Errors:
// store.Unknown - unknown error.
// store.ErrNotFound - where is no groups with provided id.
func (repo *GroupRepository) Get(ctx context.Context, id string) (*model.Group, error) {
	g := new(model.Group)

	if err := repo.pool.QueryRow(
		ctx,
		`SELECT id, "name", description, created_at, "owner" FROM groups WHERE id = $1`,
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

		repo.log.Debug(
			"unknown error while getting group",
			TraceError(err)...,
		)
		return nil, fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}
	return g, nil
}

// GetUsers ...
func (repo *GroupRepository) GetUsers(ctx context.Context, group, user string) (res []*model.UserInGroup, err error) {
	rows, err := repo.pool.Query(
		ctx,
		`SELECT u.id, u.email, uig.is_admin, r.members, r.tasks, r.reviews, r.comments
FROM user_in_group AS uig
         JOIN roles AS r on r.id = uig.role_id
         JOIN users AS u on uig.user_id = u.id
WHERE group_id = $1
  AND EXISTS(SELECT * FROM user_in_group WHERE group_id = $1 AND user_id = $2);`,
		group,
		user,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		repo.log.Error("unexpected error while doing query", TraceError(err)...)
		return nil, Unknown(err)
	}
	defer rows.Close()

	for rows.Next() {
		u := new(model.UserInGroup)
		if err = rows.Scan(&u.UserID, &u.Email, &u.IsAdmin, &u.Members, &u.Tasks, &u.Reviews, &u.Comment); err != nil {
			repo.log.Error("unknown error while scanning rows", TraceError(err)...)
			return nil, Unknown(err)
		}
		res = append(res, u)
	}

	if err = rows.Err(); err != nil {
		repo.log.Error("error occurred while reading query", TraceError(err)...)
		return nil, Unknown(err)
	}

	return
}

// UserExists ...
func (repo *GroupRepository) UserExists(ctx context.Context, group, user string) (ok bool, err error) {
	if err = repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM user_in_group WHERE group_id = $1 AND user_id = $2);`,
		group,
		user,
	).Scan(&ok); err != nil {
		repo.log.Error("get user existence in group", TraceError(err)...)
	}
	return
}

// AddUser ...
func (repo *GroupRepository) AddUser(ctx context.Context, invite string, user string) error {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		repo.log.Error("unexpected error received while starting new transaction: check drivers", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var group uuid.UUID
	var role int

	if err = tx.QueryRow(
		ctx,
		`UPDATE invites SET use_count = use_count - 1 WHERE id = $1 RETURNING group_id, role_id;`,
		invite,
	).Scan(
		&group,
		&role,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.CheckViolation {
				return store.ErrInviteIsAlreadyUsed
			}
		}
	}

	if _, err = tx.Exec(
		ctx,
		`INSERT INTO user_in_group(user_id, group_id, role_id) VALUES ($1, $2, $3);`,
		user,
		group,
		role,
	); err != nil {

		repo.log.Error("add user to group", TraceError(err)...)

		if pgErr, ok := err.(*pgconn.PgError); ok {

			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrInviteIsAlreadyUsed
			}
		}

		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	if err = tx.Commit(ctx); err != nil {
		repo.log.Error("unexpected error while doing commit transaction: check pgx driver", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	return nil
}

// AddTask creates
func (repo *GroupRepository) AddTask(ctx context.Context, task, group string) error {
	if _, err := repo.pool.Exec(ctx, `INSERT INTO task_group(task_id, group_id) VALUES ($1, $2)`, task, group); err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if !ok {
			return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
		}

		switch pgErr.Code {

		case pgerrcode.UniqueViolation:
			return store.ErrTaskAlreadyExists

		case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
			return store.ErrBadData
		}

		repo.log.Error("add task to group", TraceError(err)...)
		return fmt.Errorf("%s: %w", err.Error(), store.ErrUnknown)
	}

	return nil
}

func (repo *GroupRepository) CreateInvite(ctx context.Context, invite uuid.UUID, r *model.Role, group string, uses int) error {
	if _, err := repo.pool.Exec(ctx, `INSERT INTO invites(id, group_id, role_id, use_count)
VALUES ($1,
        $2,
        (SELECT x.id FROM roles x WHERE x.tasks = $4 AND x.members = $5 AND x.reviews = $6 AND x.comments = $7 LIMIT 1),
        $3);`, invite, group, uses, r.Tasks, r.Members, r.Reviews, r.Comments); err != nil {

		if pgErr, ok := err.(*pgconn.PgError); ok {

			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrUniqueViolation
			case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
				return store.ErrFKViolation
			}
		}
		return Unknown(err)
	}

	return nil
}
