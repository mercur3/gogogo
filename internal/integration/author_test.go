package integration

import (
	"context"
	"errors"
	"goweb/internal/common"
	"goweb/internal/db"
	"goweb/internal/handle"
	"goweb/internal/repository"
	"goweb/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	// setup
	author := db.CreateAuthorParams{
		Name: "test",
		Bio:  new("bio"),
	}

	repo := repository.New(pgPool)

	// test
	out, err := repository.WithTx(
		t.Context(),
		&repo,
		func(ctx context.Context) (db.Author, error) {
			service := service.AuthorService(&repo)
			return service.Create(ctx, author)
		},
	)

	// verify
	assert.NoError(t, err)

	svc := service.AuthorService(&repo)
	authorOut, err := svc.Get(t.Context(), out.ID)

	// transaction
	assert.NoError(t, err)
	assert.Greater(t, out.ID, int64(0))
	assert.Equal(t, author.Name, out.Name)
	assert.Equal(t, author.Bio, out.Bio)

	// reading from svc.Get
	assert.Equal(t, out.ID, authorOut.ID)
	assert.Equal(t, author.Name, authorOut.Name)
	assert.Equal(t, author.Bio, authorOut.Bio)
}

func Test_does_not_exist_returns_ErrNotFound(t *testing.T) {
	t.Parallel()

	svc := service.AuthorService(new(repository.New(pgPool)))
	_, err := svc.Get(t.Context(), -1)

	assert.ErrorIs(t, err, pgx.ErrNoRows)

	tErr, ok := errors.AsType[*common.TypedErr](err)
	assert.True(t, ok)
	assert.Equal(t, common.ErrNotFound, tErr.Kind)
}

func Test_does_not_exist_returns_404(t *testing.T) {
	t.Parallel()

	// setup
	server := handle.MakeServerFromOpenAPI(
		common.Config{MaxBodySize: 1024},
		service.AuthorService(new(repository.New(pgPool))),
		service.Book{},
	)

	req := httptest.NewRequest(http.MethodGet, "/author/-1", nil)
	w := httptest.NewRecorder()

	// test
	server.Handler.ServeHTTP(w, req)

	// verify
	assert.Equal(t, http.StatusNotFound, w.Code)
}
