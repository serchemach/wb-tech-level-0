package main

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dgraph-io/ristretto"
	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/db"
	"github.com/serchemach/wb-tech-level-0/kafka"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestServer(t *testing.T) {
	// Load the env variables because Ryuk will break if you don't
	err := godotenv.Load()
	require.NoError(t, err, "Failed to load the env variables")

	// We are not running in docker so the postgres url from .env will not work
	os.Setenv("POSTGRES_URL", "localhost:5433")

	// Setup kafka and postgres
	identifier := tc.StackIdentifier("some_ident")
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("compose_test.yaml"), identifier)
	require.NoError(t, err, "NewDockerComposeAPIWith()")

	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	require.NoError(t, compose.WithOsEnv().Up(ctx, tc.Wait(true)), "Compose no work")

	// Setup kafka and postgres connections along with cache
	dbConn, err := db.CreateDbConn()
	require.NoError(t, err, "Failed to connect to the database")

	kc, err := kafka.CreateConn()
	require.NoError(t, err, "Failed to connect to kafka")
	defer kc.Close()

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(CACHE_SIZE) * 10,
		MaxCost:     int64(CACHE_SIZE),
		BufferItems: 64,
	})
	require.NoError(t, err, "Failed to create the cache")

	err = db.InitCache(cache, dbConn, CACHE_SIZE)
	require.NoError(t, err, "Failed to initialise cache")

	go kafka.ReadTopicIndefinitely(kc, dbConn, cache)

	formFile, err := os.ReadFile("pages/form.html")
	require.NoError(t, err, "Failed to load the form file")

	t.Run("Test file retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "localhost:8080/", nil)
		w := httptest.NewRecorder()
		InterfaceHandler(w, req)

		response := w.Result()
		require.Equal(t, 200, response.StatusCode)

		buffer := make([]byte, 2000)
		n, err := response.Body.Read(buffer)
		require.NoError(t, err, "Failed to parse the file body")

		buffer = buffer[:n]
		require.Equal(t, formFile, buffer)
	})
}
