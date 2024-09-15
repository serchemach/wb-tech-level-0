package main

import (
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/serchemach/wb-tech-level-0/nats_stuff"
)

const CACHE_SIZE int = 1024

func fetchOrderWithCache(orderUid string, cache *ristretto.Cache, db *pgxpool.Pool) (*nats_stuff.Order, error) {
	orderCache, ok := cache.Get(orderUid)
	if ok {
		fmt.Printf("Cache hit for %s\n", orderUid)
		return orderCache.(*nats_stuff.Order), nil
	}
	fmt.Printf("Cache miss for %s\n", orderUid)

	order, err := fetchOrder(orderUid, db)
	if err != nil {
		return nil, err
	}

	cache.Set(orderUid, order, 1)
	cache.Wait()
	return order, nil
}

func insertOrderWithCache(order *nats_stuff.Order, cache *ristretto.Cache, db *pgxpool.Pool) error {
	cache.Set(order.OrderUid, order, 1)
	cache.Wait()
	return insertOrder(order, db)
}
