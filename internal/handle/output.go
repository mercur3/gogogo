package handle

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type serverError struct {
	msg string
}

func writeBody[T any](ctx context.Context, w http.ResponseWriter, statusCode int, body T) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.ErrorContext(ctx, "failed to write body", slog.Any("error", err))
	}
}

func setError(ctx context.Context, w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeBody(ctx, w, http.StatusNotFound, serverError{msg: err.Error()})
	} else {
		writeBody(ctx, w, http.StatusInternalServerError, serverError{msg: err.Error()})
	}
}
