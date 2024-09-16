package datamodel

import (
	"fmt"
	"time"

	"github.com/go-faker/faker/v4"
)

type Delivery struct {
	Name    string `json:"name" faker:"name"`
	Phone   string `json:"phone" faker:"phone_number"`
	Zip     string `json:"zip" faker:"word"`
	City    string `json:"city" faker:"word"`
	Address string `json:"address" faker:"real_address"`
	Region  string `json:"region" faker:"word"`
	Email   string `json:"email", faker:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction" faker:"uuid_hyphenated"`
	RequestId    string `json:"request_id" faker:"uuid_hyphenated"`
	Currency     string `json:"currency" faker:"word"`
	Provider     string `json:"provider" faker:"word"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank" faker:"word"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtId      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number" faker:"uuid_hyphenated"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name" faker:"word"`
	Sale        int    `json:"sale"`
	Size        string `json:"size" faker:"word"`
	TotalPrice  int    `json:"total_price"`
	NmId        int    `json:"nm_id"`
	Brand       string `json:"brand" faker:"word"`
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
	OrderUid          string   `json:"order_uid" faker:"uuid_hyphenated"`
	TrackNumber       string   `json:"track_number" faker:"uuid_hyphenated"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerId        string   `json:"customer_id" faker:"uuid_hyphenated"`
	DeliveryService   string   `json:"delivery_service" faker:"word"`
	Shardkey          string   `json:"shardkey"`
	SmId              int      `json:"sm_id"`
	DateCreated       Time     `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

func GenerateFakeOrder() (*Order, error) {
	a := Order{}
	err := faker.FakeData(&a)
	if err != nil {
		return nil, err
	}

	return &a, nil
}
