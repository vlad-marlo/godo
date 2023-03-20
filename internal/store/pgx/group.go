package pgx

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	if group == nil {
		return store.ErrNilReference
	}

	if err := repo.pool.QueryRow(
		ctx,
		`INSERT INTO groups(id, "name", description, "owner") VALUES ($1, $2, $3, $4) RETURNING created_at;`,
		group.ID,
		group.Name,
		group.Description,
		group.Owner,
	).Scan(&group.CreatedAt); err != nil {
		return pgError("store: group: create", err)
	}

	return nil
}

// UserExists ...
func (repo *GroupRepository) UserExists(ctx context.Context, group, user uuid.UUID) (ok bool) {
	if err := repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM user_in_group WHERE group_id = $1 AND user_id = $2);`,
		group,
		user,
	).Scan(&ok); err != nil {
		repo.log.Log(_unknownLevel, "get user existence in group", traceError(err)...)
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
		repo.log.Log(_unknownLevel, "get role of user in group", traceError(err)...)
		return nil, unknown(err)
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
		repo.log.Log(_unknownLevel, "get groups by user", traceError(err)...)
		return nil, unknown(err)
	}
	defer rows.Close()

	for rows.Next() {
		g := new(model.Group)

		if err = rows.Scan(&g.ID, &g.Name, &g.Description, &g.Owner, &g.CreatedAt); err != nil {
			repo.log.Warn("unknown error while scanning group while getting all groups by user", traceError(err)...)
			return nil, unknown(err)
		}

		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, unknown(err)
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
		repo.log.Log(_unknownLevel, "get group by id", traceError(err)...)
		return nil, unknown(err)
	}
	return g, nil

}

// GetUserIDs ...
func (repo *GroupRepository) GetUserIDs(ctx context.Context, group uuid.UUID) (ids []uuid.UUID, err error) {
	var rows pgx.Rows
	rows, err = repo.pool.Query(ctx, `SELECT uig.user_id FROM user_in_group uig WHERE uig.group_id = $1;`, group)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, store.ErrNotFound
		}
		repo.log.Log(_unknownLevel, "groups: get users", traceError(err)...)
		return nil, unknown(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			repo.log.Log(_unknownLevel, "groups: get users: scan", traceError(err)...)
			return nil, unknown(err)
		}
		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		repo.log.Log(_unknownLevel, "groups: get users: rows err", traceError(err)...)
		return nil, unknown(err)
	}

	return ids, nil
}

// TaskExists return existence of relation between task and group.
func (repo *GroupRepository) TaskExists(ctx context.Context, group, task uuid.UUID) (ok bool) {
	_ = repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT * FROM task_group tg WHERE tg.group_id = $1 AND tg.task_id = $2);`,
		group,
		task,
	).Scan(&ok)
	return
}

// AddUser adds user to group.
func (repo *GroupRepository) AddUser(ctx context.Context, roleID int32, groupID, userID uuid.UUID, isAdmin bool) error {
	if _, err := repo.pool.Exec(
		ctx,
		`INSERT INTO user_in_group(user_id, group_id, role_id, is_admin) VALUES ($1, $2, $3, $4);`,
		userID,
		groupID,
		roleID,
		isAdmin,
	); err != nil {
		return pgError("store: group: add user to group", err)
	}
	return nil
}
