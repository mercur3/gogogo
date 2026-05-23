package repository

import (
	"context"
	"goweb/internal/db"
)

type BookRepo struct {
	q *db.Queries
}

func (b *BookRepo) CreateBook(ctx context.Context, arg db.CreateBookParams) (db.Book, error) {
	return b.q.CreateBook(ctx, arg)
}

func (b *BookRepo) GetBook(ctx context.Context, id int64) (db.Book, error) {
	return b.q.GetBook(ctx, id)
}

func (b *BookRepo) PublishBook(ctx context.Context, arg db.PublishBookParams) error {
	return b.q.PublishBook(ctx, arg)
}
