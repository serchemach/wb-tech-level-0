package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/data_model"
	"github.com/serchemach/wb-tech-level-0/kafka"
)

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

	order, err := datamodel.GenerateFakeOrder(2)
	if err != nil {
		log.Fatal(err)
	}

	err = kafka.SendOrder(kafkaPartition, kafkaTopic, kafkaURL, order)

	fmt.Println(err)
}
