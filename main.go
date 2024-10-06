package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/infra/db"
	"github.com/serchemach/wb-tech-level-0/infra/kafka"
	ristrettocache "github.com/serchemach/wb-tech-level-0/service/caching"
	httptransport "github.com/serchemach/wb-tech-level-0/transport"
)

const CACHE_SIZE int = 1024

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
		os.Exit(1)
	}

	user := getEnv("POSTGRES_USER", "postgres-example-user")
	var password string
	if value, ok := os.LookupEnv("POSTGRES_PASSWORD"); !ok {
		logger.Error("There is no db user password specified in the environment variable")
		os.Exit(1)
	} else {
		password = value
	}

	url := getEnv("POSTGRES_URL", "postgres:5432")
	dbConnString := fmt.Sprintf("postgres://%s:%s@%s/wb_tech", user, password, url)

	dbConn, err := db.New(dbConnString, logger)
	if err != nil {
		logger.Error("Cannot create a connection to the database", "error", err)
		os.Exit(1)
	}

	cache, err := ristrettocache.New(logger, CACHE_SIZE, dbConn)
	if err != nil {
		logger.Error("Cannot create the cache", "error", err)
		os.Exit(1)
	}

	kafkaPartition := getEnv("KAFKA_PARTITION", "0")
	kafkaTopic := getEnv("KAFKA_TOPIC", "wb-topic")
	kafkaURL := getEnv("KAFKA_URL", "kafka:9092")

	kc, err := kafka.New(kafkaPartition, kafkaTopic, kafkaURL, logger, cache)
	if err != nil {
		logger.Error("Cannot create a connection to kafka server", "error", err)
		os.Exit(1)
	}

	kc.Listen(ctx)
	logger.Info("Successfully subscribed to the kafka topic")

	transport := httptransport.New(cache, logger)
	transport.Listen(ctx, ":8080")

	<-ctx.Done()
	// Wait until goroutines recieve the signal
	time.Sleep(time.Second)
	logger.Debug("Gracefully shut down")
}
