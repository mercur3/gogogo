package service

import (
	"context"
	"goweb/internal/db"
	"goweb/internal/repository"
)

type Author struct {
	R *repository.Repositories
}

func (a *Author) Get(ctx context.Context, id int64) (db.Author, error) {
	return a.R.Author.GetAuthor(ctx, id)
}

func (a *Author) GetAll(ctx context.Context) ([]db.Author, error) {
	return a.R.Author.ListAuthors(ctx)
}

func (a *Author) Create(ctx context.Context, params db.CreateAuthorParams) (db.Author, error) {
	return a.R.Author.CreateAuthor(ctx, params)
}
