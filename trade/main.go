package main

import (
	"flag"

	"github.com/nbasker/tools/trade/service"
)

var (
	serviceEndpoint = flag.String("service-endpoint", "localhost:8000", "Trade service endpoint")
	orderTimeout    = flag.Int("order-timeout", 10, "Order Execution Timeout")
)

func main() {
	flag.Parse()
	service.Start(*serviceEndpoint, *orderTimeout)
}
