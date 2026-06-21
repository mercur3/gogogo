package common

import "context"

type ResourceCloser interface {
	CloseResource(ctx context.Context)
}
