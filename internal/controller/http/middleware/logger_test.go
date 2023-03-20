package middleware

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWriter_Status(t *testing.T) {
	w := httptest.NewRecorder()
	wr := newLoggingRW(w)
	assert.Equal(t, wr, &responseWriter{ResponseWriter: w, statusCode: http.StatusOK})

	assert.Equal(t, wr.statusCode, wr.Status())
	assert.Equal(t, http.StatusOK, (*responseWriter)(nil).Status())
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	wr := newLoggingRW(httptest.NewRecorder())
	assert.False(t, wr.wroteHeader)

	assert.Equal(t, http.StatusOK, wr.Status())
	wr.WriteHeader(http.StatusInternalServerError)
	assert.True(t, wr.wroteHeader)

	assert.Equal(t, http.StatusInternalServerError, wr.Status())
	wr.WriteHeader(http.StatusNotFound)
	assert.True(t, wr.wroteHeader)
	assert.NotEqual(t, http.StatusNotFound, wr.Status())
}
