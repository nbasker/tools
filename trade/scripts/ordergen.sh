#!/bin/bash

for i in {1..10}
do
  transact=$(( RANDOM % 2 + 1))
  ordertype=$(( RANDOM % 2 + 1))
  quantity=$(( RANDOM % 100 + 1))
  price=$(( RANDOM % 1000 + 1))
  curl -XPOST http://localhost:8000/trade -H 'Content-Type: application/json' -d "{\"transaction\":$transact,\"quantity\":$quantity,\"price\":$price,\"order_type\":$ordertype}"
  sleep 1
done
