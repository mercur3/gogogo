package api

import (
	"context"
	"errors"
	"goweb/internal/db"
	"goweb/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	author service.Author
	book   service.Book
}

func NewServer(a service.Author, b service.Book) Server {
	return Server{author: a, book: b}
}

func (s Server) GetAllAuthors(
	ctx context.Context,
	request GetAllAuthorsRequestObject,
) (GetAllAuthorsResponseObject, error) {
	authors, err := s.author.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]Author, len(authors))
	for i, a := range authors {
		resp[i] = intoDto(a)
	}

	return GetAllAuthors200JSONResponse(resp), nil
}

func (s Server) CreateAuthor(
	ctx context.Context,
	request CreateAuthorRequestObject,
) (CreateAuthorResponseObject, error) {
	author, err := s.author.Create(ctx, db.CreateAuthorParams{
		Name: request.Body.Name,
		Bio:  request.Body.Bio,
	})
	if err != nil {
		return nil, err
	}

	return CreateAuthor201JSONResponse(intoDto(author)), nil
}

func (s Server) GetAuthor(
	ctx context.Context,
	request GetAuthorRequestObject,
) (GetAuthorResponseObject, error) {
	author, err := s.author.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return GetAuthor200JSONResponse(intoDto(author)), nil
}

func (s Server) UpdateAuthor(
	ctx context.Context,
	request UpdateAuthorRequestObject,
) (UpdateAuthorResponseObject, error) {
	return nil, s.author.Update(ctx, db.UpdateAuthorParams{
		ID:   request.Id,
		Name: request.Body.Name,
		Bio:  request.Body.Bio,
	})
}

func (s Server) CreateBook(
	ctx context.Context,
	request CreateBookRequestObject,
) (CreateBookResponseObject, error) {
	book, err := s.book.CreateBook(ctx, db.CreateBookParams{
		Title:       request.Body.Title,
		PublishedAt: request.Body.PublishedAt,
	})
	if err != nil {
		return CreateBook400JSONResponse(ErrorMsg{
			Msg:       err.Error(),
			RequestId: uuid.New(),
		}), nil
	}

	return CreateBook200JSONResponse(Book{
		Id:          book.ID,
		Title:       book.Title,
		PublishedAt: book.PublishedAt,
	}), nil
}

func (s Server) PublishBook(
	ctx context.Context,
	request PublishBookRequestObject,
) (PublishBookResponseObject, error) {
	if err := s.book.PublishBook(ctx, request.BookId, request.Params.AuthorId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PublishBook404JSONResponse(ErrorMsg{
				Msg:       err.Error(),
				RequestId: uuid.New(),
			}), nil
		}
		return PublishBook400JSONResponse(ErrorMsg{
			Msg:       err.Error(),
			RequestId: uuid.New(),
		}), nil
	}

	return PublishBook204Response{}, nil
}

func (s Server) GetBook(
	ctx context.Context,
	request GetBookRequestObject,
) (GetBookResponseObject, error) {
	book, err := s.book.GetBook(ctx, request.Id)
	if err != nil {
		return GetBook404JSONResponse(ErrorMsg{
			Msg:       err.Error(),
			RequestId: uuid.New(),
		}), nil
	}

	return GetBook200JSONResponse(Book{
		Id:          book.ID,
		Title:       book.Title,
		PublishedAt: book.PublishedAt,
	}), nil
}

func intoDto(a db.Author) Author {
	return Author{
		Id:   a.ID,
		Name: a.Name,
		Bio:  a.Bio,
	}
}
