package service

import (
	"os"

	"github.com/nbasker/tools/trade/api"
	"github.com/nbasker/tools/trade/matcher"
	"github.com/nbasker/tools/trade/store"

	"github.com/sirupsen/logrus"
)

// Start the service.
func Start(srvEp string, oTimeout int) {
	log := logrus.New()
	log.Out = os.Stdout

	log.Debug("service.Start()")

	orders := make(chan matcher.Order)
	complete := make(chan matcher.Order)

	store := store.NewStorageService(complete, log)
	go store.StoreCompletedOrders()

	match := matcher.NewMatcherService(orders, oTimeout, log)
	go match.ExecuteOrders()

	serve := api.NewApiService(srvEp, orders, store, log)
	serve.Run()
}
