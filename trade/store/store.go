package store

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nbasker/tools/trade/matcher"
)

// Retriever fetches executed orders from a storage or database
type Store interface {
	// StoreCompletedOrders persists the executed and timedout orders
	StoreCompletedOrders()

	// RetrieveExecutedOrders gets the executed orders from store.
	RetrieveExecutedOrders() string
}

// storageService persists and retrieves completed orders
type storageService struct {
	complete <-chan *matcher.Order
	log      *logrus.Logger
	store    map[string]*matcher.Order
}

// NewMatcherService instantiates order matching service
func NewStorageService(
	complete <-chan *matcher.Order,
	log *logrus.Logger,
) Store {
	return &storageService{
		complete: complete,
		log:      log,
		store:    make(map[string]*matcher.Order),
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
			}).Debug("Persist")
			s.store[o.Id.String()] = o
		}
	}
}

// ReceiveOrders gets the orders from a channel and stores in memory
func (s *storageService) RetrieveExecutedOrders() string {
	s.log.Info("Retrieving completed orders (executed and timedout)")
	oResp := "Id/Time => [Buy/Sell, Price, Placed, Executed, Left, Status]\n"
	for _, o := range s.store {
		s.log.WithFields(logrus.Fields{
			"OrderId":     o.Id.String(),
			"Transaction": o.Transaction,
			"OrderType":   o.OrderType,
			"Quantity":    o.Quantity,
			"Executed":    o.Executed,
			"Price":       o.Price,
			"OrderTime":   o.OrderTime.Format(time.UnixDate),
		}).Debug("Processed")
		oResp += fmt.Sprintf("%s/%s => [ %s, %d, %d, %d, %d, %s ]\n",
			o.Id.String(),
			o.OrderTime.Format(time.UnixDate),
			o.Transaction.String(),
			o.Price,
			o.PlacedQuantity,
			o.Executed,
			o.Quantity,
			o.Status.String())
	}
	return oResp
}
