package nats_stuff

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/stan.go"
)

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestId    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtId      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmId        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Time time.Time

func (d *Time) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str != "" && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// parse string
	t, err := time.Parse("2006-01-02T15:04:05-0700", str)
	if err == nil {
		*d = Time(t)
		return nil
	}
	return fmt.Errorf("invalid duration type %T, value: '%s'", b, b)
}

func (d Time) MarshalJSON() ([]byte, error) {
	string := fmt.Sprintf("\"%s\"", time.Time(d).Format("2006-01-02T15:04:05-0700"))
	return []byte(string), nil
}

type Order struct {
	OrderUid          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerId        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	Shardkey          string   `json:"shardkey"`
	SmId              int      `json:"sm_id"`
	DateCreated       Time     `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func CreateConn() (stan.Conn, error) {
	clusterId := getEnv("STAN_CLUSTER_ID", "test-cluster")
	clientId := getEnv("STAN_CLIENT_ID", "wb-tech-level-0")
	stanURL := getEnv("STAN_URL", "nats://localhost:4222")

	return stan.Connect(clusterId, clientId, stan.NatsURL(stanURL))
}
