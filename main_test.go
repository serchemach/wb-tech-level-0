package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/joho/godotenv"
	datamodel "github.com/serchemach/wb-tech-level-0/data_model"
	"github.com/serchemach/wb-tech-level-0/db"
	"github.com/serchemach/wb-tech-level-0/kafka"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func TestServer(t *testing.T) {
	// Load the env variables because Ryuk will break if you don't
	err := godotenv.Load()
	require.NoError(t, err, "Failed to load the env variables")

	// We are not running in docker so the postgres and kafka url from .env will not work
	os.Setenv("POSTGRES_URL", "localhost:5433")
	os.Setenv("KAFKA_URL", "localhost:9094")

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

	testOrder, err := datamodel.GenerateFakeOrder(2)
	require.NoError(t, err, "Error while generating a fake order")

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

	t.Run("Test empty order id retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "localhost:8080/api/v1/order?order_uid=", nil)
		w := httptest.NewRecorder()
		OrderHandler(w, req, cache, dbConn)

		response := w.Result()
		require.NotEqual(t, 200, response.StatusCode)
	})

	t.Run("Test wrong order id retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "localhost:8080/api/v1/order?order_uid=1111", nil)
		w := httptest.NewRecorder()
		OrderHandler(w, req, cache, dbConn)

		response := w.Result()
		require.NotEqual(t, 200, response.StatusCode)
	})

	t.Run("Test fake order sending", func(t *testing.T) {
		kafkaPartition, _ := strconv.Atoi(getEnv("KAFKA_PARTITION", "0"))
		kafkaTopic := getEnv("KAFKA_TOPIC", "wb-topic")
		kafkaURL := "localhost:9094"
		err = kafka.SendOrder(kafkaPartition, kafkaTopic, kafkaURL, testOrder)
		require.NoError(t, err, "Error while sending a fake order")
	})

	// Wait for the order to propagate to the server
	time.Sleep(1 * time.Second)

	t.Run("Test correct order id retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("localhost:8080/api/v1/order?order_uid=%s", testOrder.OrderUid), nil)
		w := httptest.NewRecorder()
		OrderHandler(w, req, cache, dbConn)

		response := w.Result()
		require.Equal(t, 200, response.StatusCode)

		buffer := make([]byte, 20000)
		n, err := response.Body.Read(buffer)
		require.NoError(t, err, "Failed to parse the file body")

		order, err := kafka.ParseMessage(buffer[:n])
		require.NoError(t, err, "Failed to parse the recieved order")
		require.Equal(t, testOrder, order, "Order from the server is not equal to the one sent")
	})
}
