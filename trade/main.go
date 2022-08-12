package main

import (
	"fmt"

	"github.com/nbasker/tools/trade/service"
)

func main() {
	fmt.Println("A trading server.")
	service.Start()
}
