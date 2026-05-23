package repository

import (
	"context"
	"goweb/internal/db"
)

type AuthorRepo struct {
	q *db.Queries
}

func (a *AuthorRepo) CreateAuthor(
	ctx context.Context,
	arg db.CreateAuthorParams,
) (db.Author, error) {
	return a.q.CreateAuthor(ctx, arg)
}

func (a *AuthorRepo) DeleteAuthor(ctx context.Context, id int64) error {
	return a.q.DeleteAuthor(ctx, id)
}

func (a *AuthorRepo) GetAuthor(ctx context.Context, id int64) (db.Author, error) {
	return a.q.GetAuthor(ctx, id)
}

func (a *AuthorRepo) ListAuthors(ctx context.Context) ([]db.Author, error) {
	return a.q.ListAuthors(ctx)
}

func (a *AuthorRepo) UpdateAuthor(ctx context.Context, arg db.UpdateAuthorParams) error {
	return a.q.UpdateAuthor(ctx, arg)
}
