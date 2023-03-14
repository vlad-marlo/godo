package pgx

import (
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
)

// traceError checks if error is PgError and return fields that shows more info about pg error.
//
// If error is not Pg then will return just zap.Error field with error in it.
func traceError(err error) []zap.Field {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
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
		zap.L().DPanic("unknown error", fields...)
	default:
	}
	return fields
}

// unknown wraps store.ErrUnknown with provided error.
func unknown(err error) error {
	return fmt.Errorf("%w: %s", store.ErrUnknown, err.Error())
}

func pgError(msg string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return store.ErrUniqueViolation
		case pgerrcode.ForeignKeyViolation, pgerrcode.InvalidForeignKey:
			return store.ErrFKViolation
		}
	}
	zap.L().Log(_unknownLevel, msg, traceError(err)...)

	return unknown(err)
}
