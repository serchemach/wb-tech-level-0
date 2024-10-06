package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/serchemach/wb-tech-level-0/data_model"
)

type Store interface {
	AddOrder(*datamodel.Order) error
	GetOrder(orderUid string) (*datamodel.Order, error)
}

type KafkaInfra interface {
	Listen(context.Context)
}

type kafkaInfra struct {
	reader *kafka.Reader
	store  Store
	logger *slog.Logger
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (k *kafkaInfra) Listen(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				k.logger.Debug("Gracefully shut down kafka reader")
				k.reader.Close()
				return
			default:
			}

			m, err := k.reader.ReadMessage(ctx)
			if err != nil {
				// fmt.Printf("Error while reading the message: %s\n", err)
				continue
			}

			order, err := datamodel.ParseOrder(m.Value)
			if err != nil {
				k.logger.Error("Error while converting the data recieved from the channel", "error", err)
				continue
			}

			k.logger.Info("Recieved an order", "value", order)
			err = k.store.AddOrder(order)
			if err != nil {
				k.logger.Error("Error while inserting the order into the database", "error", err)
			}
		}
	}()
}

func New(partition, topic, url string, logger *slog.Logger, store Store) (KafkaInfra, error) {
	kafkaPartition, err := strconv.Atoi(partition)
	if err != nil {
		return nil, err
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Topic:     topic,
		Partition: kafkaPartition,
		Brokers:   []string{url},
	})
	return &kafkaInfra{
		reader: reader,
		logger: logger,
		store:  store,
	}, nil
}

func SendOrder(kafkaPartition int, kafkaTopic string, kafkaURL string, order *datamodel.Order) error {
	kc, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, kafkaTopic, kafkaPartition)
	if err != nil {
		return err
	}
	defer kc.Close()

	encodedOrder, err := json.Marshal(order)
	if err != nil {
		return err
	}

	kc.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = kc.WriteMessages(
		kafka.Message{Value: encodedOrder},
	)
	if err != nil {
		return err
	}

	return nil
}
