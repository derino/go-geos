package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	var (
		dbName   = "testdb"
		user     = "testuser"
		password = "testpassword"
	)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := pool.Run("postgis/postgis", "13-3.1-alpine", []string{
		"POSTGRES_DB=" + dbName,
		"POSTGRES_PASSWORD=" + password,
		"POSTGRES_USER=" + user,
	})
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, pool.Purge(resource))
	}()

	var db *sql.DB
	require.NoError(t, pool.Retry(func() error {
		dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?binary_parameters=yes&sslmode=disable", user, password, resource.GetPort("5432/tcp"), dbName)
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return db.Ping()
	}))

	require.NoError(t, createDB(db))

	require.NoError(t, populateDB(db))

	r := bytes.NewBufferString(`{"name":"Paris","geometry":{"type":"Point","coordinates":[2.3508,48.8567]}}`)
	require.NoError(t, readGeoJSON(db, r))

	w := &strings.Builder{}
	require.NoError(t, writeGeoJSON(db, w))
	assert.Equal(t, strings.Join([]string{
		`{"id":1,"name":"London","geometry":{"type":"Point","coordinates":[0.1275,51.50722]}}`,
		`{"id":2,"name":"Berlin","geometry":{"type":"Point","coordinates":[13.405,52.52]}}`,
		`{"id":3,"name":"Paris","geometry":{"type":"Point","coordinates":[2.3508,48.8567]}}`,
	}, "\n")+"\n", w.String())
}
