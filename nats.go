package main

import (
	"github.com/nats-io/stan.go"
	"os"
	"time"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func createConn() (stan.Conn, error) {
	clusterId := getEnv("STAN_CLUSTER_ID", "test-cluster")
	clientId := getEnv("STAN_CLIENT_ID", "wb-tech-level-0")
	stanURL := getEnv("STAN_URL", "nats://nats-streaming:4222")

	return stan.Connect(clusterId, clientId, stan.NatsURL(stanURL))
}

func messageHandler(m *stan.Msg) {

}

func createSubscription(sc stan.Conn) (stan.Subscription, error) {
	channelName := getEnv("STAN_CHANNEL_NAME", "wb-channel")

	return sc.Subscribe(channelName, messageHandler, stan.StartWithLastReceived(), stan.AckWait(20*time.Second))
}
