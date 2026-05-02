package handle

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type serverError struct {
	msg string
}

func writeBody[T any](w http.ResponseWriter, statusCode int, body T) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to write body", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func setError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeBody(w, http.StatusNotFound, serverError{msg: err.Error()})
	} else {
		writeBody(w, http.StatusInternalServerError, serverError{msg: err.Error()})
	}
}
