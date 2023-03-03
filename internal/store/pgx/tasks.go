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

//
//func (repo *TaskRepository) Create(ctx context.Context, task *model.Task) error {
//	if _, err := repo.pool.Exec(
//		ctx,
//		`INSERT INTO tasks(id, "name", description, created_at, created_by) VALUES ($1, $2, $3, $4, $5);`,
//		task.ID, task.Name, task.Description, task.CreatedAt, task.CreatedBy,
//	); err != nil {
//		if pgErr, ok := err.(*pgconn.PgError); ok {
//			switch pgErr.Code {
//			case pgerrcode.UniqueViolation:
//				return store.ErrUniqueViolation
//			}
//		}
//
//		repo.log.Error("create task", TraceError(err)...)
//		return Unknown(err)
//	}
//	return nil
//}
//
//func (repo *TaskRepository) AddToGroupUsers(ctx context.Context, group, task string) error {
//	if _, err := repo.pool.Exec(
//		ctx,
//		`INSERT INTO task_user(user_id, task_id)
//VALUES ((SELECT uig.user_id FROM user_in_group uig WHERE group_id = $1), $2) ON CONFLICT DO NOTHING;`,
//		group,
//		task,
//	); err != nil {
//		if pgErr, ok := err.(*pgconn.PgError); ok {
//			switch pgErr.Code {
//			case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
//				return store.ErrFKViolation
//			}
//		}
//		repo.log.Error("")
//		return Unknown(err)
//	}
//	return nil
//}

// GetByGroup ...
func (repo *TaskRepository) GetByGroup(ctx context.Context, group uuid.UUID) ([]*model.Task, error) {
	q := `
SELECT t.id, t.name, t.description, t.created_at, t.created_by, t.status
FROM tasks t
         JOIN task_group tg on t.id = tg.task_id
WHERE tg.group_id = $1`

	rows, err := repo.pool.Query(ctx, q, group)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}

		return nil, Unknown(err)
	}

	defer rows.Close()

	var resp []*model.Task
	for rows.Next() {
		t := new(model.Task)

		if err = rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.CreatedBy, &t.Status); err != nil {
			repo.log.Warn("scan task while getting group tasks", TraceError(err)...)
			return nil, Unknown(err)
		}

		resp = append(resp, t)
	}

	return resp, nil
}
