package service

import (
	"context"
	"goweb/internal/db"
	"goweb/internal/repository"
)

type Author struct {
	r *repository.Repositories
}

func AuthorService(r *repository.Repositories) Author {
	return Author{r: r}
}

func (a *Author) Get(ctx context.Context, id int64) (db.Author, error) {
	return a.r.Author.GetAuthor(ctx, id)
}

func (a *Author) GetAll(ctx context.Context) ([]db.Author, error) {
	return a.r.Author.ListAuthors(ctx)
}

func (a *Author) Create(ctx context.Context, params db.CreateAuthorParams) (db.Author, error) {
	return a.r.Author.CreateAuthor(ctx, params)
}
