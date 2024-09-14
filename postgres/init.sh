#!/bin/bash
psql -U $POSTGRES_USER -c 'create database wb_tech;'

# Create the tables and add some mock data
psql -v ON_ERROR_STOP=1 -U $POSTGRES_USER -d wb_tech  <<-EOSQL
  CREATE SCHEMA order_scheme;

  CREATE TABLE order_scheme.delivery (
    id serial primary key,
    name text not null,
    phone text not null,
    zip text not null,
    city text not null,
    address text not null,
    region text not null,
    email text not null
  ); 

  CREATE TABLE order_scheme.payment (
    transaction text primary key,
    request_id text not null,
    currency text not null,
    provider text not null,
    amount integer not null,
    payment_dt integer not null,
    bank text not null,
    delivery_cost integer not null,
    goods_total integer not null,
    custom_fee integer not null
  );

  CREATE TABLE order_scheme.item (
    chrt_id integer primary key,
    track_number text not null,
    price integer not null,
    rid text not null,
    name text not null,
    sale integer not null,
    size text not null,
    total_price integer not null,
    nm_id integer not null,
    brand text not null,
    status integer not null
  );

  CREATE TABLE order_scheme.order (
    order_uid  text primary key,
    track_number text not null,
    entry text not null,
    delivery_id integer not null,
    payment_id text not null, 
    locale text not null,
    internal_signature text not null,
    customer_id text not null,
    delivery_service text not null,
    shardkey text not null,
    sm_id integer not null,
    date_created timestamp not null,
    oof_shard text not null,
    FOREIGN KEY (delivery_id) REFERENCES order_scheme.delivery(id),
    FOREIGN KEY (payment_id) REFERENCES order_scheme.payment(transaction)
  );

  CREATE TABLE order_scheme.order_item_conn (
    order_uid text not null,
    chrt_id integer not null,
    PRIMARY KEY (order_uid, chrt_id),
    FOREIGN KEY (order_uid) REFERENCES order_scheme.order(order_uid),
    FOREIGN KEY (chrt_id) REFERENCES order_scheme.item(chrt_id)
  );
     
EOSQL
