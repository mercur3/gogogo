package service

import (
	"context"
	"goweb/internal/db"
	"goweb/internal/repository"
)

type Book struct {
	r *repository.Repositories
}

func BookService(r *repository.Repositories) Book {
	return Book{r}
}

func (b *Book) PublishBook(ctx context.Context, bookID int64, authorID int64) error {
	_, err := repository.WithTx(ctx, b.r, func() (struct{}, error) {
		if _, err := b.r.Book.GetBook(ctx, bookID); err != nil {
			return struct{}{}, err
		}

		if _, err := b.r.Author.GetAuthor(ctx, authorID); err != nil {
			return struct{}{}, err
		}

		err := b.r.Book.PublishBook(ctx, db.PublishBookParams{BookID: bookID, AuthorID: authorID})
		return struct{}{}, err
	})

	return err
}
