package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	// "html"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error loading .env file: %s\n", err))
	}

	db, err := createDbConn()
	if err != nil {
		log.Fatal("Cannot create a connection to the database: ", err)
	}

	sc, err := nats_stuff.CreateConn()
	if err != nil {
		log.Fatal("Cannot create a connection to nats-streaming server: ", err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(CACHE_SIZE) * 10,
		MaxCost:     int64(CACHE_SIZE),
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal("Error while creating the cache: ", err)
	}

	err = initCache(cache, db)
	if err != nil {
		log.Fatal("Error while initialising the cache: ", err)
	}

	fmt.Printf("Added %d orders to cache\n", cache.Metrics.CostAdded())
	fmt.Printf("Cache max cost %d\n", cache.MaxCost())

	sub, err := CreateSubscription(sc, db, cache)
	if err != nil {
		log.Fatal("Error while subscribing to the channel: ", err)
	}
	defer sub.Close()

	fmt.Println("Successfully subscribed to the nats-streaming server")

	http.HandleFunc("GET /order", func(w http.ResponseWriter, r *http.Request) {
		orderUid := r.URL.Query().Get("order_uid")
		if orderUid == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No order id given"))
			return
		}

		order, err := fetchOrderWithCache(orderUid, cache, db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error while fetching the order data: %s", err)))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(order)
		fmt.Printf("Error while encoding the json: %s\n", err)
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		htmlFile, err := os.ReadFile("pages/form.html")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Println(err)
		}

		w.Write(htmlFile)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
