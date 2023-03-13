package pgx

import (
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/vlad-marlo/godo/internal/store"
	"go.uber.org/zap"
	"testing"
)

func TestPgError(t *testing.T) {
	tt := []struct {
		name   string
		err    error
		expect error
	}{
		{"unknown", errors.New(""), store.ErrUnknown},
		{"invalid FK", &pgconn.PgError{Code: pgerrcode.InvalidForeignKey}, store.ErrFKViolation},
		{"FK violation", &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}, store.ErrFKViolation},
		{"unique violation", &pgconn.PgError{Code: pgerrcode.UniqueViolation}, store.ErrUniqueViolation},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := pgError("", tc.err)
			assert.ErrorIs(t, err, tc.expect)
		})
	}
}

func TestTraceErr_PanicSeverity(t *testing.T) {
	zap.ReplaceGlobals(zap.NewNop())
	TraceError(&pgconn.PgError{Severity: "PANIC"})
}
