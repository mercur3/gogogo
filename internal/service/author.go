package service

import (
	"context"
	"goweb/internal/db"
	"goweb/internal/otel"
	"goweb/internal/repository"
)

const authorService = "author-service"

type Author struct {
	r *repository.Repositories
}

func AuthorService(r *repository.Repositories) Author {
	return Author{r: r}
}

func (a *Author) Get(ctx context.Context, id int64) (db.Author, error) {
	ctx, span := otel.Tracer().Start(ctx, authorService+":get")
	defer span.End()

	author, err := a.r.Author.GetAuthor(ctx, id)
	if err != nil {
		otel.SetError(span, err)
		return db.Author{}, err
	}

	return author, nil
}

func (a *Author) GetAll(ctx context.Context) ([]db.Author, error) {
	return a.r.Author.ListAuthors(ctx)
}

func (a *Author) Create(ctx context.Context, params db.CreateAuthorParams) (db.Author, error) {
	return a.r.Author.CreateAuthor(ctx, params)
}

func (a *Author) Update(ctx context.Context, params db.UpdateAuthorParams) error {
	ctx, span := otel.Tracer().Start(ctx, authorService+":update")
	defer span.End()

	if err := a.r.Author.UpdateAuthor(ctx, params); err != nil {
		otel.SetError(span, err)
		return err
	}
	return nil
}
