package pgx

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

// TraceError checks if error is PgError and return fields that shows more info about pg error.
//
// If error is not Pg then will return just zap.Error field with error in it.
func TraceError(err error) []zap.Field {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return []zap.Field{zap.Error(err)}
	}
	fields := []zap.Field{
		zap.String("code", pgErr.Code),
		zap.String("message", pgErr.Message),
		zap.String("data type name", pgErr.DataTypeName),
		zap.String("hint", pgErr.Hint),
		zap.String("routine", pgErr.Routine),
		zap.String("constraint", pgErr.ConstraintName),
		zap.String("column", pgErr.ColumnName),
		zap.String("detail", pgErr.Detail),
		zap.Error(pgErr),
	}
	switch pgErr.Severity {
	case "PANIC":
		zap.L().Panic("unknown error", fields...)
	default:
	}
	return fields
}

// Unknown wraps store.ErrUnknown with provided error.
func Unknown(err error) error {
	return fmt.Errorf("%w: %s", store.ErrUnknown, err.Error())
}
