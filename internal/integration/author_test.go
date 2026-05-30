package integration

import (
	"context"
	"goweb/internal/db"
	"goweb/internal/repository"
	"goweb/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	author := db.CreateAuthorParams{
		Name: "test",
		Bio:  new("bio"),
	}

	repo := repository.New(pgPool)
	out, err := repository.WithTx(
		t.Context(),
		&repo,
		func(ctx context.Context) (db.Author, error) {
			service := service.AuthorService(&repo)
			return service.Create(ctx, author)
		},
	)

	assert.NoError(t, err)
	assert.Greater(t, out.ID, int64(0))
	assert.Equal(t, author.Name, out.Name)
	assert.Equal(t, author.Bio, out.Bio)
}
