package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	// "time"

	"github.com/dgraph-io/ristretto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/serchemach/wb-tech-level-0/data_model"
	"github.com/serchemach/wb-tech-level-0/db"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func CreateConn() (*kafka.Reader, error) {
	kafkaPartition, err := strconv.Atoi(getEnv("KAFKA_PARTITION", "0"))
	if err != nil {
		return nil, err
	}
	kafkaTopic := getEnv("KAFKA_TOPIC", "wb-topic")
	kafkaURL := getEnv("KAFKA_URL", "kafka:9092")

	fmt.Println(kafkaURL)

	// return kafka.DialLeader(context.Background(), "tcp", kafkaURL, kafkaTopic, kafkaPartition)
	return kafka.NewReader(kafka.ReaderConfig{
		Topic:     kafkaTopic,
		Partition: kafkaPartition,
		Brokers:   []string{kafkaURL},
	}), nil
}

func ReadTopicIndefinitely(kafkaConn *kafka.Reader, dbConn *pgxpool.Pool, cache *ristretto.Cache) {
	// kafkaConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	// batch := kafkaConn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	// buffer := make([]byte, 10e3) // 10KB max per message

	for {
		m, err := kafkaConn.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Error while reading the message: %s\n", err)
			continue
		}
		order, err := ParseMessage(m.Value)
		if err != nil {
			fmt.Printf("Error while converting the data recieved from the channel: %s\n", err)
		}

		fmt.Printf("Recieved an order: %v\n", order)
		err = db.InsertOrderWithCache(order, cache, dbConn)
		if err != nil {
			fmt.Printf("Error while inserting the order into the database: %s\n", err)
		}
	}
}

func ParseMessage(m []byte) (*datamodel.Order, error) {
	fmt.Printf("Recieved a message: %s\n", m)
	var order datamodel.Order
	decoder := json.NewDecoder(bytes.NewReader(m))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func SendOrder(kafkaPartition int, kafkaTopic string, kafkaURL string, order *datamodel.Order) error {
	kc, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, kafkaTopic, kafkaPartition)
	if err != nil {
		return err
	}
	defer kc.Close()

	encodedOrder, err := json.Marshal(order)
	if err != nil {
		return err
	}

	kc.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = kc.WriteMessages(
		kafka.Message{Value: encodedOrder},
	)
	if err != nil {
		return err
	}

	return nil
}
