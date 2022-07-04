package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithApplicationName(t *testing.T) {
	param := "ApplicationName"
	var conf Config
	f := WithApplicationName(param)
	f(&conf)
	require.Equal(t, conf.ApplicationName, param)
}

func TestWithDatabase(t *testing.T) {
	param := "Database"
	var conf Config
	f := WithDatabase(param)
	f(&conf)
	require.Equal(t, conf.Database, param)
}

func TestWithDialTimeout(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithDialTimeout(param)
	f(&conf)
	require.Equal(t, conf.DialTimeout, param)
}

func TestWithHealthCheckResultTTL(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithHealthCheckResultTTL(param)
	f(&conf)
	require.Equal(t, conf.HealthCheckResultTTL, param)
}

func TestWithHealthCheckTableName(t *testing.T) {
	param := "HealthCheckTableName"
	var conf Config
	f := WithHealthCheckTableName(param)
	f(&conf)
	require.Equal(t, conf.HealthCheckTableName, param)
}

func TestWithHost(t *testing.T) {
	param := "Host"
	var conf Config
	f := WithHost(param)
	f(&conf)
	require.Equal(t, conf.Host, param)
}

func TestWithIdleCheckFrequency(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithIdleCheckFrequency(param)
	f(&conf)
	require.Equal(t, conf.IdleCheckFrequency, param)
}

func TestWithIdleTimeout(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithIdleTimeout(param)
	f(&conf)
	require.Equal(t, conf.IdleTimeout, param)
}

func TestWithMaxConnAge(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithMaxConnAge(param)
	f(&conf)
	require.Equal(t, conf.MaxConnAge, param)
}

func TestWithMaxRetries(t *testing.T) {
	param := 42
	var conf Config
	f := WithMaxRetries(param)
	f(&conf)
	require.Equal(t, conf.MaxRetries, param)
}

func TestWithMaxRetryBackoff(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithMaxRetryBackoff(param)
	f(&conf)
	require.Equal(t, conf.MaxRetryBackoff, param)
}

func TestWithMinIdleConns(t *testing.T) {
	param := 42
	var conf Config
	f := WithMinIdleConns(param)
	f(&conf)
	require.Equal(t, conf.MinIdleConns, param)
}

func TestWithMinRetryBackoff(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithMinRetryBackoff(param)
	f(&conf)
	require.Equal(t, conf.MinRetryBackoff, param)
}

func TestWithPassword(t *testing.T) {
	param := "Password"
	var conf Config
	f := WithPassword(param)
	f(&conf)
	require.Equal(t, conf.Password, param)
}

func TestWithPoolSize(t *testing.T) {
	param := 42
	var conf Config
	f := WithPoolSize(param)
	f(&conf)
	require.Equal(t, conf.PoolSize, param)
}

func TestWithPoolTimeout(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithPoolTimeout(param)
	f(&conf)
	require.Equal(t, conf.PoolTimeout, param)
}

func TestWithPort(t *testing.T) {
	param := 42
	var conf Config
	f := WithPort(param)
	f(&conf)
	require.Equal(t, conf.Port, param)
}

func TestWithReadTimeout(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithReadTimeout(param)
	f(&conf)
	require.Equal(t, conf.ReadTimeout, param)
}

func TestWithRetryStatementTimeout(t *testing.T) {
	param := true
	var conf Config
	f := WithRetryStatementTimeout(param)
	f(&conf)
	require.Equal(t, conf.RetryStatementTimeout, param)
}

func TestWithUser(t *testing.T) {
	param := "User"
	var conf Config
	f := WithUser(param)
	f(&conf)
	require.Equal(t, conf.User, param)
}

func TestWithWriteTimeout(t *testing.T) {
	param := 5 * time.Second
	var conf Config
	f := WithWriteTimeout(param)
	f(&conf)
	require.Equal(t, conf.WriteTimeout, param)
}

func TestWithLogReadWriteOnly(t *testing.T) {
	cases := [][]bool{
		{
			true, true,
		},
		{
			false, true,
		},
		{
			true, false,
		},
		{
			false, false,
		},
	}
	for _, tc := range cases {
		read := tc[0]
		write := tc[1]
		var conf Config
		f := WithQueryLogging(read, write)
		f(&conf)
		assert.Equal(t, conf.LogRead, read)
		assert.Equal(t, conf.LogWrite, write)
	}
}
