package common

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func Test_Is_works_with_Parent(t *testing.T) {
	t.Parallel()

	err := NewTypedErr(ErrNotFound, "", pgx.ErrNoRows)
	assert.True(t, errors.Is(err, pgx.ErrNoRows))
}

func Test_AsType_works(t *testing.T) {
	t.Parallel()

	var err error = NewTypedErr(ErrAlreadyExists, "", pgx.ErrTooManyRows)
	out, ok := errors.AsType[*TypedErr](err)

	assert.True(t, ok)
	assert.Equal(t, out.Kind, ErrKind(ErrAlreadyExists))
	assert.Equal(t, out.Msg, "")
	assert.Equal(t, out.Parent, pgx.ErrTooManyRows)
}
