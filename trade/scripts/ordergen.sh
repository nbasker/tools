#!/bin/bash

for i in {1..10}
do
  transact=$(gshuf -i 1-2 -n 1)
  ordertype=$(gshuf -i 1-2 -n 1)
  quantity=$(gshuf -i 10-50 -n 1)
  price=$(gshuf -i 511-516 -n 1)
  curl -XPOST http://localhost:8000/trade -H 'Content-Type: application/json' -d "{\"transaction\":$transact,\"quantity\":$quantity,\"price\":$price,\"order_type\":$ordertype}"
  sleep 1
done
