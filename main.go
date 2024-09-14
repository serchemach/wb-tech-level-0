package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/nats_stuff"
	"html"
	"log"
	"net/http"
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

	sub, err := nats_stuff.CreateSubscription(sc)
	if err != nil {
		log.Fatal("Error while subscribing to the channel: ", err)
	}
	defer sub.Close()

	fmt.Println("Successfully subscribed to the nats-streaming server")

	http.HandleFunc("GET /order", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
