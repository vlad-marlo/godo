package pgx

import (
	"github.com/jackc/pgx/v5/pgxpool"
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
