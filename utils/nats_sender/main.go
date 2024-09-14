package main

import (
	"encoding/json"
	"fmt"
	"log"
	// "time"

	"github.com/bxcodec/faker/v3"
	"github.com/joho/godotenv"
	"github.com/serchemach/wb-tech-level-0/nats_stuff"
)

// func generateOrder() nats_stuff.Order {

// }

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Printf("Error loading .env file: %v\n", err))
	}
	sc, err := nats_stuff.CreateConn()
	if err != nil {
		log.Fatal("Cannot create a connection to nats-streaming server: ", err)
	}

	testOrder := nats_stuff.Order{}
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

	err = sc.Publish("wb-channel", encodedOrder)
	fmt.Println(err)
}
