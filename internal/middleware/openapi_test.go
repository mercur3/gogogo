package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_x_request_id_not_set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requestId string
	}{
		{name: "not set by client", requestId: ""},
		{name: "set to invalid value", requestId: "1234-4321"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set(requestIdHeader, tt.requestId)
			w := httptest.NewRecorder()

			// test
			ctx, requestId := setRequestId(t.Context(), r, w)

			// verify
			assert.NotEqual(t, tt.requestId, requestId)
			assert.Equal(t, requestId, ctx.Value(RequestID{}).(uuid.UUID).String())
		})
	}
}

func Test_x_request_id_set_to_correct_value(t *testing.T) {
	t.Parallel()

	// setup
	requestId := uuid.NewString()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(requestIdHeader, requestId)
	w := httptest.NewRecorder()

	// test
	ctx, requestIdOut := setRequestId(t.Context(), r, w)

	// verify
	assert.Equal(t, requestId, requestIdOut)
	assert.Equal(t, requestId, ctx.Value(RequestID{}).(uuid.UUID).String())
}
