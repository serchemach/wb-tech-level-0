package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/serchemach/wb-tech-level-0/nats_stuff"
)

type DeliveryDb struct {
	Id      int    `db:"id"`
	Name    string `db:"name"`
	Phone   string `db:"phone"`
	Zip     string `db:"zip"`
	City    string `db:"city"`
	Address string `db:"address"`
	Region  string `db:"region"`
	Email   string `db:"email"`
}

type PaymentDb struct {
	Transaction  string `db:"transaction"`
	RequestId    string `db:"request_id"`
	Currency     string `db:"currency"`
	Provider     string `db:"provider"`
	Amount       int    `db:"amount"`
	PaymentDt    int    `db:"payment_dt"`
	Bank         string `db:"bank"`
	DeliveryCost int    `db:"delivery_cost"`
	GoodsTotal   int    `db:"goods_total"`
	CustomFee    int    `db:"custom_fee"`
}

type ItemDb struct {
	ChrtId      int    `db:"chrt_id"`
	TrackNumber string `db:"track_number"`
	Price       int    `db:"price"`
	Rid         string `db:"rid"`
	Name        string `db:"name"`
	Sale        int    `db:"sale"`
	Size        string `db:"size"`
	TotalPrice  int    `db:"total_price"`
	NmId        int    `db:"nm_id"`
	Brand       string `db:"brand"`
	Status      int    `db:"status"`
}

type OrderDb struct {
	OrderUid          string    `db:"order_uid"`
	TrackNumber       string    `db:"track_number"`
	Entry             string    `db:"entry"`
	DeliveryId        int       `db:"delivery_id"`
	PaymentId         string    `db:"payment_id"`
	Locale            string    `db:"locale"`
	InternalSignature string    `db:"internal_signature"`
	CustomerId        string    `db:"customer_id"`
	DeliveryService   string    `db:"delivery_service"`
	Shardkey          string    `db:"shardkey"`
	SmId              int       `db:"sm_id"`
	DateCreated       time.Time `db:"date_created"`
	OofShard          string    `db:"oof_shard"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type NoPasswordError struct{}

func (m *NoPasswordError) Error() string {
	return "There is no db user password specified in the environment variable"
}

func createDbConn() (*pgxpool.Pool, error) {
	user := getEnv("POSTGRES_USER", "postgres-example-user")
	var password string
	if value, ok := os.LookupEnv("POSTGRES_PASSWORD"); !ok {
		return nil, &NoPasswordError{}
	} else {
		password = value
	}

	connString := fmt.Sprintf("postgres://%s:%s@postgres:5432/wb_tech", user, password)

	return pgxpool.New(context.Background(), connString)
}

func insertOrder(order *nats_stuff.Order, conn *pgxpool.Pool) error {
	queryString := fmt.Sprintf("INSERT INTO order_scheme.delivery (name, phone, zip, city, address, region, email) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s') RETURNING id;", order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	var deliveryId int
	row := conn.QueryRow(context.Background(), queryString)
	err := row.Scan(&deliveryId)
	if err != nil {
		return err
	}

	queryString = fmt.Sprintf("INSERT INTO order_scheme.payment VALUES ('%s', '%s', '%s', '%s', %d, %d, '%s', %d, %d, %d)", order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	_, err = conn.Exec(context.Background(), queryString)
	if err != nil {
		return err
	}

	queryString = fmt.Sprintf("INSERT INTO order_scheme.order VALUES ('%s', '%s', '%s', %d, '%s', '%s', '%s', '%s', '%s', '%s', %d, '%s', '%s')", order.OrderUid, order.TrackNumber, order.Entry, deliveryId, order.Payment.Transaction, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, time.Time(order.DateCreated).Format("2006-01-02T15:04:05-0700"), order.OofShard)
	_, err = conn.Exec(context.Background(), queryString)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		queryString = fmt.Sprintf("INSERT INTO order_scheme.item VALUES (%d, '%s', %d, '%s', '%s', %d, '%s', %d, %d, '%s', %d)", item.ChrtId, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status)
		_, err = conn.Exec(context.Background(), queryString)
		if err != nil {
			return err
		}

		queryString = fmt.Sprintf("INSERT INTO order_scheme.order_item_conn VALUES ('%s', %d)", order.OrderUid, item.ChrtId)
		_, err = conn.Exec(context.Background(), queryString)
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchOrder(orderUid string, conn *pgxpool.Pool) (*nats_stuff.Order, error) {
	queryString := fmt.Sprintf("SELECT * FROM order_scheme.order WHERE order_uid = '%s'", orderUid)
	rows, _ := conn.Query(context.Background(), queryString)
	orderDb, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[OrderDb])
	if err != nil {
		return nil, err
	}

	queryString = fmt.Sprintf("SELECT * FROM order_scheme.delivery WHERE id = %d", orderDb.DeliveryId)
	rows, _ = conn.Query(context.Background(), queryString)
	deliveryDb, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[DeliveryDb])
	if err != nil {
		return nil, err
	}

	queryString = fmt.Sprintf("SELECT * FROM order_scheme.payment WHERE transaction = '%s'", orderDb.PaymentId)
	rows, _ = conn.Query(context.Background(), queryString)
	paymentDb, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[PaymentDb])
	if err != nil {
		return nil, err
	}

	queryString = fmt.Sprintf("SELECT order_scheme.item.chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM order_scheme.item INNER JOIN order_scheme.order_item_conn ON order_scheme.order_item_conn.chrt_id = order_scheme.item.chrt_id WHERE order_uid = '%s'", orderDb.OrderUid)
	rows, _ = conn.Query(context.Background(), queryString)
	itemsDb, err := pgx.CollectRows(rows, pgx.RowToStructByName[ItemDb])
	if err != nil {
		return nil, err
	}

	delivery := nats_stuff.Delivery{
		Name:    deliveryDb.Name,
		Phone:   deliveryDb.Phone,
		Zip:     deliveryDb.Zip,
		City:    deliveryDb.City,
		Address: deliveryDb.Address,
		Region:  deliveryDb.Region,
		Email:   deliveryDb.Email,
	}

	payment := nats_stuff.Payment{
		Transaction:  paymentDb.Transaction,
		RequestId:    paymentDb.RequestId,
		Currency:     paymentDb.Currency,
		Provider:     paymentDb.Provider,
		Amount:       paymentDb.Amount,
		PaymentDt:    paymentDb.PaymentDt,
		Bank:         paymentDb.Bank,
		DeliveryCost: paymentDb.DeliveryCost,
		GoodsTotal:   paymentDb.GoodsTotal,
		CustomFee:    paymentDb.CustomFee,
	}

	items := make([]nats_stuff.Item, len(itemsDb))
	for i := 0; i < len(itemsDb); i++ {
		items[i] = nats_stuff.Item{
			ChrtId:      itemsDb[i].ChrtId,
			TrackNumber: itemsDb[i].TrackNumber,
			Price:       itemsDb[i].Price,
			Rid:         itemsDb[i].Rid,
			Name:        itemsDb[i].Name,
			Sale:        itemsDb[i].Sale,
			Size:        itemsDb[i].Size,
			TotalPrice:  itemsDb[i].TotalPrice,
			NmId:        itemsDb[i].NmId,
			Brand:       itemsDb[i].Brand,
			Status:      itemsDb[i].Status,
		}
	}

	return &nats_stuff.Order{
		OrderUid:          orderDb.OrderUid,
		TrackNumber:       orderDb.TrackNumber,
		Entry:             orderDb.Entry,
		Delivery:          delivery,
		Payment:           payment,
		Items:             items,
		Locale:            orderDb.Locale,
		InternalSignature: orderDb.InternalSignature,
		CustomerId:        orderDb.CustomerId,
		DeliveryService:   orderDb.DeliveryService,
		Shardkey:          orderDb.Shardkey,
		SmId:              orderDb.SmId,
		DateCreated:       nats_stuff.Time(orderDb.DateCreated),
		OofShard:          orderDb.OofShard,
	}, nil
}
