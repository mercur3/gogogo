package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDbInitialization(t *testing.T) {
	t.Parallel()

	row := pgPool.QueryRow(context.Background(), "select 1")
	output := ""

	err := row.Scan(&output)
	assert.NoError(t, err)
	assert.Equal(t, "1", output)
}
