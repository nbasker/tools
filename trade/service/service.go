package service

import (
	"fmt"

	"github.com/nbasker/tools/trade/api"
	"github.com/nbasker/tools/trade/matcher"
)

// Start the service.
func Start() {
	fmt.Println("service.Start()")
	api.Start()
	matcher.Start()
}
