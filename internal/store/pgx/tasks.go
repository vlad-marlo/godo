package pgx

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vlad-marlo/godo/internal/model"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

var _ store.TaskRepository = (*TaskRepository)(nil)

type TaskRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewTaskRepository return newly initialized object of task repository
func NewTaskRepository(cli Client) *TaskRepository {
	return &TaskRepository{
		pool: cli.P(),
		log:  cli.L(),
	}
}

// AllByUser return all tasks related to user.
//
// Task will be returned to user if this cases:
// * User is admin of group to which task is related;
// * user is related to group;
// * user has permission to read tasks in group where task is created.
func (repo *TaskRepository) AllByUser(ctx context.Context, user uuid.UUID) ([]*model.Task, error) {
	q := `SELECT t.id, t.name, t.description, t.created_at, t.created_by, t.status
FROM tasks t
         LEFT JOIN task_user tu on t.id = tu.task_id
         LEFT JOIN task_group tg on t.id = tg.task_id
         LEFT JOIN user_in_group uig on tg.group_id = uig.group_id
         LEFT JOIN roles r on r.id = uig.role_id
WHERE tu.user_id = $1 OR t.created_by = $1 OR (uig.user_id = $1 AND (uig.is_admin OR r.tasks >= 2));`

	rows, err := repo.pool.Query(ctx, q, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		repo.log.Log(_unknownLevel, "get tasks by group and user", TraceError(err)...)

		return nil, Unknown(err)
	}

	defer rows.Close()

	var resp []*model.Task
	for rows.Next() {
		t := new(model.Task)

		if err = rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.CreatedBy, &t.Status); err != nil {
			repo.log.Log(_unknownLevel, "scan task while getting group tasks", TraceError(err)...)
			return nil, Unknown(err)
		}

		resp = append(resp, t)
	}
	if err = rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		return nil, Unknown(err)
	}

	return resp, nil
}

// AllByGroupAndUser return all related to
func (repo *TaskRepository) AllByGroupAndUser(ctx context.Context, group uuid.UUID, user uuid.UUID) ([]*model.Task, error) {
	// данный вопрос возвращает все задачи, к которым относится пользователь - он администратор группы, имеет право на чтение, или указан как получатель задачи.
	q := `SELECT t.id, t.name, t.description, t.created_at, t.created_by, t.status
FROM tasks t
         JOIN task_group tg on t.id = tg.task_id
         LEFT JOIN task_user tu on t.id = tu.task_id
         JOIN user_in_group uig on tu.user_id = uig.user_id
         JOIN roles r on uig.role_id = r.id
WHERE tg.group_id = $1
  AND (tu.user_id = $2 OR (uig.user_id = $2 AND (uig.is_admin OR r.tasks >= 1)));`

	rows, err := repo.pool.Query(ctx, q, group, user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		repo.log.Log(_unknownLevel, "get tasks by group and user", TraceError(err)...)

		return nil, Unknown(err)
	}

	defer rows.Close()

	var resp []*model.Task
	for rows.Next() {
		t := new(model.Task)

		if err = rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.CreatedBy, &t.Status); err != nil {
			repo.log.Log(_unknownLevel, "scan task while getting group tasks", TraceError(err)...)
			return nil, Unknown(err)
		}

		resp = append(resp, t)
	}

	return resp, nil
}

// GetByUserAndID return tasks if it related to user.
//
// Task will be returned to user if this cases:
// * User is admin of group to which task is related;
// * user is related to group;
// * user has permission to read tasks in group where task is created.
func (repo *TaskRepository) GetByUserAndID(ctx context.Context, user, task uuid.UUID) (*model.Task, error) {
	q := `SELECT t.id, t.name, t.description, t.created_at, t.created_by, t.status
FROM tasks t
         LEFT JOIN task_user tu on t.id = tu.task_id
         LEFT JOIN task_group tg on t.id = tg.task_id
         LEFT JOIN user_in_group uig on tg.group_id = uig.group_id
         LEFT JOIN roles r on r.id = uig.role_id
WHERE t.id = $2 AND (tu.user_id = $1 OR t.created_by = $1 OR (uig.user_id = $1 AND (uig.is_admin OR r.tasks >= 2)));`
	t := new(model.Task)

	if err := repo.pool.QueryRow(ctx, q, user, task).Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.CreatedBy, &t.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		repo.log.Log(_unknownLevel, "get tasks by group and user", TraceError(err)...)

		return nil, Unknown(err)
	}
	return t, nil
}

func (repo *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	if _, err := repo.pool.Exec(
		ctx,
		`INSERT INTO tasks(id, name, description, created_at, created_by, status)
VALUES ($1, $2, $3, $4, $5, $6);`,
		task.ID,
		task.Name,
		task.Description,
		task.CreatedAt,
		task.CreatedBy,
		task.Status,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrUniqueViolation
			case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
				return store.ErrFKViolation
			}
		}
		repo.log.Log(_unknownLevel, "tasks: create", TraceError(err)...)

		return Unknown(err)
	}

	return nil
}

func (repo *TaskRepository) AddToGroup(ctx context.Context, task, group uuid.UUID) error {
	if _, err := repo.pool.Exec(ctx, `INSERT INTO task_group(task_id, group_id) VALUES ($1, $2);`, task, group); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrUniqueViolation
			case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
				return store.ErrFKViolation
			}
		}

		repo.log.Log(_unknownLevel, "add task to group", TraceError(err)...)
		return Unknown(err)
	}
	return nil
}

// AddToUser add
func (repo *TaskRepository) AddToUser(ctx context.Context, from, task, to uuid.UUID) error {
	var ok bool
	// check existence of group where exists both of users and user, who want to add task has permission.
	_ = repo.pool.QueryRow(
		ctx,
		`SELECT EXISTS(
               SELECT *
               FROM groups g
                        LEFT JOIN user_in_group uig on g.id = uig.group_id
                        LEFT JOIN roles r on r.id = uig.role_id
               WHERE EXISTS(SELECT * FROM user_in_group WHERE user_id = $1 AND group_id = g.id)
                 AND EXISTS(SELECT * FROM user_in_group WHERE user_id = $2 AND group_id = g.id)
                 AND (uig = $2 AND (uig.is_admin OR r.tasks >= 2))
           );`,
		to,
		from,
	).Scan(&ok)
	if !ok {
		return nil
	}

	if _, err := repo.pool.Exec(ctx, `INSERT INTO task_user(user_id, task_id)
VALUES ($1, $2);`, to, task, from); err != nil {
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
