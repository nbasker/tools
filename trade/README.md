## Trade
The trade is a sample program to receive simple buy and sell requests and match them.
The trade order is made of the following inputs.
* Transaction: buy or sell
* PlacedQuantity
* Price
* OrderType: market or limit

The order status clarified using the following additional fields.
* Status: placed or completed or timedout
* ExecutedQuantity
* OrderTime
* UUID: to uniquely identify the transaction within the system

### Directory Structure

- `api`: A basic REST API interface to place and get order status.
- `matcher`: A order matching logic implementation.
- `store`: A store for the completed, timedout or cancelled orders.
- `service`: A glue that ties all the packages together

```
trade/
├─ api/
│  ├─ api.go
│  ├─ api_test.go
├─ matcher/
│  ├─ matcher.go
│  ├─ matcher_test.go
├─ store/
│  ├─ store.go
├─ service/
│  ├─ service.go
├─ scripts/
│  ├─ ordergen.sh
│  ├─ getorder.sh
├─ main.go
├─ README.md
```

### Run Instructions

The help options available are
```
./trade --help
Usage of ./trade:
  -order-timeout int
        Order Execution Timeout (default 10)
  -service-endpoint string
        Trade service endpoint (default "localhost:8000")
```

Execution procedure is
```
go build && ./trade 
INFO[0000] Starting REST Api Service                     endpoint="localhost:8000"
INFO[0000] Starting to collected completed orders and persist 
INFO[0000] Starting to Execute Orders                    OrderTimeout=10
```

### References
1. Go Web Server Skeleton https://betterprogramming.pub/implementing-a-basic-http-server-using-go-a59b1888359b
2. JSON Encoder https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
3. Go maps https://go.dev/blog/maps
