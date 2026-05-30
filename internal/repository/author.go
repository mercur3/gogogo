package repository

import (
	"context"
	"errors"
	"goweb/internal/common"
	"goweb/internal/db"

	"github.com/jackc/pgx/v5"
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
	count, err := a.q.DeleteAuthor(ctx, id)
	if err != nil {
		return common.NewTypedErr(common.ErrUnknown, "Failed to delete author", err)
	} else if count == 0 {
		return common.NewTypedErr(common.ErrNotFound, "Does not exist", err)
	}

	return nil
}

func (a *AuthorRepo) GetAuthor(ctx context.Context, id int64) (db.Author, error) {
	author, err := a.q.GetAuthor(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Author{}, common.NewTypedErr(common.ErrNotFound, "User not found", err)
		} else {
			return db.Author{}, common.NewTypedErr(
				common.ErrUnknown,
				"Failed to get user data",
				err,
			)
		}
	}

	return author, nil
}

func (a *AuthorRepo) ListAuthors(ctx context.Context) ([]db.Author, error) {
	authors, err := a.q.ListAuthors(ctx)
	if err != nil {
		return []db.Author{}, common.NewTypedErr(common.ErrUnknown, "Failed to get authors", err)
	}

	return authors, nil
}

func (a *AuthorRepo) UpdateAuthor(ctx context.Context, arg db.UpdateAuthorParams) error {
	count, err := a.q.UpdateAuthor(ctx, arg)
	if err != nil {
		return common.NewTypedErr(common.ErrUnknown, "Failed to update author", err)
	} else if count == 0 {
		return common.NewTypedErr(common.ErrNotFound, "Does not exist", err)
	}

	return nil
}
