package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_requiredEnv_return_err_when_missing(t *testing.T) {
	env := "TEST_ENV"
	assert.NoError(t, os.Unsetenv(env))
	assert.ErrorContains(t, requiredEnv(env, nil, nil), "missing required env: ")
}

func Test_ParseConfgis_parses_all_envs(t *testing.T) {
	// setup
	assert.NoError(t, os.Setenv(maxBodySizeEnv, "1234"))
	assert.NoError(t, os.Setenv(dbUserEnv, "db-user"))
	assert.NoError(t, os.Setenv(dbPasswordEnv, "db-pass"))
	assert.NoError(t, os.Setenv(dbHostEnv, "127.0.0.1"))
	assert.NoError(t, os.Setenv(dbPortEnv, "9999"))
	assert.NoError(t, os.Setenv(dbNameEnv, "db-name"))

	// test
	cfg, err := ParseConfigs()
	assert.NoError(t, err)
	assert.Equal(t, int64(1234), cfg.MaxBodySize)
	assert.Equal(t, "db-user", cfg.DbUser)
	assert.Equal(t, "db-pass", cfg.DbPassword)
	assert.Equal(t, "127.0.0.1", cfg.DbHost)
	assert.Equal(t, "9999", cfg.DbPort)
	assert.Equal(t, "db-name", cfg.DbName)
}
