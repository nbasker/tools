## Trade
The trade is a sample program to receive simple buy and sell requests and match them.
The trade order is made of the following inputs.
* Transaction: buy or sell
* PlacedQuantity
* Price
* OrderType: market or limit

The order status provided using the following additional fields.
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

### Design
<img width="664" alt="Trade_DesignDiagram" src="https://user-images.githubusercontent.com/16254163/184537116-9b75c9f9-f574-4547-95d9-fd02cdae4fdf.png">

The above diagram shows the high level design and message flow of the system.
* The API is a net/http based webserver that receives external requests and places them on order write-only channel.
* The Matcher module is supplied with order (read-only) channel and complete (write-only) channel. It receives orders from order-channel and stores buy orders in buyMap and sell orders in sellMap. The "price" of the order is the key for the map. It matches the buy and sell orders based on price. It takes a timeout parameter and checks for timedout orders. The completed and timedout orders are removed and sent on complete channel.
* The store module is given complete (read-only) channel. It receives the executed orders and stores them in DB (currently only an in memory map). It exposes a Retrieve() interface to fetch orders stored in the DB.

### Additional Design Considerations
1. Use of bufferred channels and further examine the size of the buffer-channel so that matcher has a way to store data if it is over-whelmed by requests.
2. Investigate on how to scale matcher. It is currently running one goroutine. Need to further look if sharding is possible or if a distributed memory store such as memcached or redis would help.
3. Remove all logging and put on debug mode.
4. Enable debug hooks so that the in memory data structure can be dumped for investigation purposes.
5. A clean way to close channel and exit.
6. Currently store uses a map without lock as it can be read while being written. This needs to be fixed. A database would help as it can store data for future analysis as well.
7. The Matcher is cleaning up the orders if they go beyond a time. The clean currently loops and can become expensive. Need to investigate a way to optimize it.
8. Test is only on API and matcher and only on a few sample functions. This needs to be enhanced for better code coverage.

### Testing Strategy

For unit testing, using the go test framework. A sample testing output provided. Further function level unit testing needs to be added. Additionally coverage for storage and API packages have to be implemented.

```
go test -count=1 -v ./...
?       github.com/nbasker/tools/trade  [no test files]
?       github.com/nbasker/tools/trade/api      [no test files]
=== RUN   Test_Matcher_ExecuteOrders
=== RUN   Test_Matcher_ExecuteOrders/MatchingPrice,_QuantityEqual
=== RUN   Test_Matcher_ExecuteOrders/MatchingPrice,_QuantityUnEqual
=== RUN   Test_Matcher_ExecuteOrders/NonMatchingPrice
--- PASS: Test_Matcher_ExecuteOrders (10.00s)
    --- PASS: Test_Matcher_ExecuteOrders/MatchingPrice,_QuantityEqual (0.00s)
    --- PASS: Test_Matcher_ExecuteOrders/MatchingPrice,_QuantityUnEqual (5.00s)
    --- PASS: Test_Matcher_ExecuteOrders/NonMatchingPrice (5.00s)
PASS
ok      github.com/nbasker/tools/trade/matcher  10.116s
?       github.com/nbasker/tools/trade/service  [no test files]
?       github.com/nbasker/tools/trade/store    [no test files]
```

A second strategy is using scripts that can generate random orders to simulate real life situations. Multiple instances of these can be run. The final order executed status from the database can be queried and validated.

### References
1. Go Web Server Skeleton https://betterprogramming.pub/implementing-a-basic-http-server-using-go-a59b1888359b
2. JSON Encoder https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
3. Go maps https://go.dev/blog/maps
