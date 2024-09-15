package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/stan.go"
	"github.com/serchemach/wb-tech-level-0/nats_stuff"
)

func messageHandler(m *stan.Msg, dbConn *pgxpool.Pool, cache *ristretto.Cache) {
	fmt.Printf("Recieved a message: %s\n", m.Data)
	var order nats_stuff.Order
	decoder := json.NewDecoder(bytes.NewReader(m.Data))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&order)
	if err != nil {
		fmt.Printf("Error while converting the data recieved from the channel: %s\n", err)
		return
	}

	fmt.Printf("Recieved an order: %v\n", order)
	err = insertOrderWithCache(&order, cache, dbConn)
	if err != nil {
		fmt.Printf("Error while inserting the order into the database: %s\n", err)
		return
	}
}

func CreateSubscription(sc stan.Conn, dbConn *pgxpool.Pool, cache *ristretto.Cache) (stan.Subscription, error) {
	channelName := getEnv("STAN_CHANNEL_NAME", "wb-channel")

	return sc.Subscribe(channelName, func(m *stan.Msg) { messageHandler(m, dbConn, cache) }, stan.StartWithLastReceived(), stan.AckWait(20*time.Second))
}
