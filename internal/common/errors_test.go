package common

import (
	"errors"
	"os"
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

func Test_Error_prints_only_msg_when_no_parent(t *testing.T) {
	t.Parallel()

	msg := "root error"
	err := NewTypedErr(ErrUnknown, msg, nil)

	assert.Equal(t, msg, err.Error())
}

func Test_Error_prints_both_msg_and_parent_when_not_nil(t *testing.T) {
	t.Parallel()

	err := NewTypedErr(ErrUnknown, "root error", os.ErrClosed)

	assert.Equal(t, "root error: file already closed", err.Error())
}
