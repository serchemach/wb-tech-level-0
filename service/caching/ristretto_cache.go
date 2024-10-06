package ristrettocache

import (
	"log/slog"

	"github.com/dgraph-io/ristretto"
	datamodel "github.com/serchemach/wb-tech-level-0/data_model"
)

type RistrettoPassthrough interface {
	AddOrder(*datamodel.Order) error
	GetOrder(orderUid string) (*datamodel.Order, error)
}

type DBInfra interface {
	AddOrder(*datamodel.Order) error
	GetOrder(orderUid string) (*datamodel.Order, error)
	GetLastOrderIds(int) ([]string, error)
}

type ristrettoCache struct {
	cache   *ristretto.Cache
	logger  *slog.Logger
	dbInfra DBInfra
	size    int
}

func (c *ristrettoCache) init() error {
	orderIds, err := c.dbInfra.GetLastOrderIds(c.size)
	if err != nil {
		return err
	}

	for _, orderId := range orderIds {
		cur_order, err := c.dbInfra.GetOrder(orderId)
		if err != nil {
			return err
		}
		ok := c.cache.Set(orderId, cur_order, 1)
		c.logger.Debug("Tried setting cache", "key", orderId, "result", ok)
		c.cache.Wait()
	}

	return nil
}

func (c *ristrettoCache) GetOrder(orderUid string) (*datamodel.Order, error) {
	orderCache, ok := c.cache.Get(orderUid)
	if ok {
		c.logger.Info("Cache hit", "key", orderUid)
		return orderCache.(*datamodel.Order), nil
	}
	c.logger.Info("Cache miss", "key", orderUid)

	order, err := c.dbInfra.GetOrder(orderUid)
	if err != nil {
		return nil, err
	}

	c.cache.Set(orderUid, order, 1)
	c.cache.Wait()
	return order, nil
}

func (c *ristrettoCache) AddOrder(order *datamodel.Order) error {
	c.cache.Set(order.OrderUid, order, 1)
	c.cache.Wait()
	return c.dbInfra.AddOrder(order)
}

func New(logger *slog.Logger, cacheSize int, db DBInfra) (RistrettoPassthrough, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(cacheSize) * 10,
		MaxCost:     int64(cacheSize),
		BufferItems: 64,
	})

	if err != nil {
		return nil, err
	}

	store := ristrettoCache{
		cache:   cache,
		size:    cacheSize,
		logger:  logger,
		dbInfra: db,
	}

	err = store.init()
	return &store, err
}
