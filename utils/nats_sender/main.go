package main

import (
	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/nats_stuff"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: %v\n", err)
	}
	sc, err := nats_stuff.CreateConn()
	if err != nil {
		log.Fatal("Cannot create a connection to nats-streaming server: ", err)
	}

	err = sc.Publish("wb-channel", []byte("Hello channel!"))
}
