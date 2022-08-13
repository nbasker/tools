package store

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nbasker/tools/trade/matcher"
)

// Retriever fetches executed orders from a storage or database
type Store interface {
	// StoreCompletedOrders persists the executed and timedout orders
	StoreCompletedOrders()

	// RetrieveExecutedOrders gets the executed orders from store.
	RetrieveExecutedOrders()
}

// storageService persists and retrieves completed orders
type storageService struct {
	complete <-chan matcher.Order
	log      *logrus.Logger
	store    map[string]matcher.Order
}

// NewMatcherService instantiates order matching service
func NewStorageService(
	complete <-chan matcher.Order,
	log *logrus.Logger,
) Store {
	return &storageService{
		complete: complete,
		log:      log,
		store:    make(map[string]matcher.Order),
	}
}

// StoreCompletedOrders persists the executed and timed out orders
func (s *storageService) StoreCompletedOrders() {
	s.log.Info("Starting to collected completed orders and persist")
	for {
		select {
		case o := <-s.complete:
			s.log.WithFields(logrus.Fields{
				"OrderId":     o.Id.String(),
				"Transaction": o.Transaction,
				"OrderType":   o.OrderType,
				"Quantity":    o.Quantity,
				"Executed":    o.Executed,
				"Price":       o.Price,
				"OrderTime":   o.OrderTime.Format(time.UnixDate),
			}).Info("Persist")
		}
	}
}

// ReceiveOrders gets the orders from a channel and stores in memory
func (s *storageService) RetrieveExecutedOrders() {
	s.log.Info("Retrieving completed orders (executed and timedout)")
}
