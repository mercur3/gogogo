package handle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goweb/internal/service"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// real http server
// all the middlewares need to be configured manually
func TestCallingV1GetAll_AlwaysWorks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(getAllHandler))
	defer server.Close()
	client := server.Client()

	var tests = []struct {
		name string
		body io.Reader
	}{
		{"empty", nil},
		{"dummy", bytes.NewBufferString("dummy")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+"/v1/test", tt.body)
			assert.NoError(t, err)
			// req := httptest.NewRequest(http.MethodGet, server.URL+"/v1/test", tt.body)
			// w := httptest.NewRecorder()

			res, err := client.Do(req)
			assert.NoError(t, err)

			// assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %s", err)
				}
			}()
			var body string
			// err := json.NewDecoder(w.Body).Decode(&body)
			assert.NoError(t, json.NewDecoder(res.Body).Decode(&body))
			assert.Equal(t, "GET /v1/test", string(body))
		})
	}
}

// handler test
// the middlewares are already configured when making the server
func TestGetById_FailsWhenNan(t *testing.T) {
	t.Parallel()

	// setup
	server := MakeServer(service.Author{})

	req := httptest.NewRequest(http.MethodGet, "/v1/test/abc", nil)
	w := httptest.NewRecorder()

	// test
	server.Handler.ServeHTTP(w, req)

	// verify
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, w.Body)
}

func TestGetById_WorksWhenNum(t *testing.T) {
	t.Parallel()

	server := MakeServer(service.Author{})

	for _, tt := range []int{-1, 0, 1} {
		t.Run(fmt.Sprintf("param = %d", tt), func(t *testing.T) {
			strValue := strconv.Itoa(tt)
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/test/%s", strValue), nil)
			w := httptest.NewRecorder()

			server.Handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			var body string
			assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
			assert.Equal(t, fmt.Sprintf("GET /v1/test/{%s}", strValue), body)
		})
	}
}
