#!/bin/bash
psql -U $POSTGRES_USER -c 'create database wb-tech;'

# Create the tables and add some mock data
psql -v ON_ERROR_STOP=1 -U $POSTGRES_USER -d sample  <<-EOSQL
     
EOSQL
