package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	grafanalgtm "github.com/testcontainers/testcontainers-go/modules/grafana-lgtm"
)

func Test_connection_works(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	container, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	defer func() {
		assert.NoError(t, testcontainers.TerminateContainer(container))
	}()
	require.NoError(t, err)

	closer, err := InitOtel(ctx, container.MustOtlpGrpcEndpoint(ctx))
	defer closer.CloseResource(ctx)
	assert.NoError(t, err)
}
