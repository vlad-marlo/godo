package pgx

import (
	"github.com/jackc/pgx/v5/pgconn"
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
	return []zap.Field{
		zap.String("code", pgErr.Code),
		zap.String("message", pgErr.Message),
		zap.String("data type name", pgErr.DataTypeName),
		zap.String("hint", pgErr.Hint),
		zap.Error(pgErr),
	}
}