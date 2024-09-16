package main

import (
	"encoding/json"
	"fmt"
	"os"

	"log"
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/db"
	"github.com/serchemach/wb-tech-level-0/kafka"
)

const CACHE_SIZE int = 1024

func OrderHandler(w http.ResponseWriter, r *http.Request, cache *ristretto.Cache, dbConn *pgxpool.Pool) {
	orderUid := r.URL.Query().Get("order_uid")
	if orderUid == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No order id given"))
		return
	}

	order, err := db.FetchOrderWithCache(orderUid, cache, dbConn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while fetching the order data: %s", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(order)
	fmt.Printf("Error while encoding the json: %s\n", err)
}

func InterfaceHandler(w http.ResponseWriter, r *http.Request) {
	htmlFile, err := os.ReadFile("pages/form.html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(err)
	}

	w.Write(htmlFile)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error loading .env file: %s\n", err))
	}

	dbConn, err := db.CreateDbConn()
	if err != nil {
		log.Fatal("Cannot create a connection to the database: ", err)
	}

	kc, err := kafka.CreateConn()
	if err != nil {
		log.Fatal("Cannot create a connection to kafka server: ", err)
	}
	defer kc.Close()

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(CACHE_SIZE) * 10,
		MaxCost:     int64(CACHE_SIZE),
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal("Error while creating the cache: ", err)
	}

	err = db.InitCache(cache, dbConn, CACHE_SIZE)
	if err != nil {
		log.Fatal("Error while initialising the cache: ", err)
	}

	fmt.Printf("Added %d orders to cache\n", cache.Metrics.CostAdded())
	fmt.Printf("Cache max cost %d\n", cache.MaxCost())

	go kafka.ReadTopicIndefinitely(kc, dbConn, cache)
	fmt.Println("Successfully subscribed to the kafka topic")

	http.HandleFunc("GET /order", func(w http.ResponseWriter, r *http.Request) { OrderHandler(w, r, cache, dbConn) })
	http.HandleFunc("GET /", InterfaceHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
