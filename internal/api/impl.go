package api

import (
	"context"
	"goweb/internal/db"
	"goweb/internal/service"

	"github.com/google/uuid"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	author service.Author
}

func NewServer(a service.Author) Server {
	return Server{author: a}
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
		return CreateAuthor400JSONResponse{
			Msg:       err.Error(),
			RequestId: uuid.New(),
		}, nil
	}

	return CreateAuthor200JSONResponse(intoDto(author)), nil
}

func (s Server) GetAuthor(
	ctx context.Context,
	request GetAuthorRequestObject,
) (GetAuthorResponseObject, error) {
	author, err := s.author.Get(ctx, request.Id)
	if err != nil {
		return GetAuthor404JSONResponse{
			Msg:       err.Error(),
			RequestId: uuid.New(),
		}, nil
	}

	return GetAuthor200JSONResponse(intoDto(author)), nil
}

func intoDto(a db.Author) Author {
	return Author{
		Id:   a.ID,
		Name: a.Name,
		Bio:  a.Bio,
	}
}
