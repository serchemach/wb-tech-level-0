package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	// "time"

	"github.com/bxcodec/faker/v3"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/serchemach/wb-tech-level-0/data_model"
)

// func generateOrder() nats_stuff.Order {

// }

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Printf("Error loading .env file: %v\n", err))
	}
	kafkaPartition, _ := strconv.Atoi(getEnv("KAFKA_PARTITION", "0"))
	kafkaTopic := getEnv("KAFKA_TOPIC", "wb-topic")
	kafkaURL := getEnv("KAFKA_URL", "localhost:9094")

	kc, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, kafkaTopic, kafkaPartition)
	if err != nil {
		log.Fatal("Cannot create a connection to kafka server: ", err)
	}
	defer kc.Close()

	testOrder := datamodel.Order{}
	err = faker.FakeData(&testOrder)
	if err != nil {
		fmt.Println(err)
	}
	if len(testOrder.Items) > 1 {
		testOrder.Items = testOrder.Items[:2]
	}
	fmt.Printf("%+v", testOrder)

	encodedOrder, err := json.Marshal(testOrder)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(string(encodedOrder))
	// tTime, err := time.Parse("2006-01-02T15:04:05-0700", "2014-05-16T08:28:06.801064-04:00")
	// testOrder.DateCreated = nats_stuff.Time(tTime)

	kc.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = kc.WriteMessages(
		kafka.Message{Value: encodedOrder},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	fmt.Println(err)
}
