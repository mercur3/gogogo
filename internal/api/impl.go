package api

import (
	"context"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct{}

func (s Server) GetPing(ctx context.Context, request GetPingRequestObject) (GetPingResponseObject, error) {
	resp := GetPing200JSONResponse{Ping: "pong"}
	return resp, nil
}
