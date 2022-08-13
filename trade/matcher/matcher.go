package matcher

import (
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Transaction enum
type Transaction int

const (
	Buy Transaction = iota + 1
	Sell
)

func (t Transaction) String() string {
	switch t {
	case Buy:
		return "buy"
	case Sell:
		return "sell"
	}
	return "unknown"
}

// OrderType enum
type OrderType int

const (
	Market OrderType = iota + 1
	Limit
)

func (t OrderType) String() string {
	switch t {
	case Market:
		return "market"
	case Limit:
		return "limit"
	}
	return "unknown"
}

// Order defines the order placed for trade
type Order struct {
	Id          uuid.UUID   `json:"id,omitempty"`
	OrderTime   time.Time   `json:"order_time,omitempty"`
	Transaction Transaction `json:"transaction,omitempty"`
	Quantity    int         `json:"quantity,omitempty"`
	Executed    int         `json:"executed,omitempty"`
	Price       int         `json:"price,omitempty"`
	OrderType   OrderType   `json:"order_type,omitempty"`
}

// Matcher that receives orders and executes
type Matcher interface {
	// Execute Orders matches the buy and sell order from in memory maps.
	ExecuteOrders()
}

// matcherService implements the order processing
type matcherService struct {
	och      <-chan Order
	oTimeout int
	log      *logrus.Logger
	buy      map[int][]Order
	sell     map[int][]Order
}

// NewMatcherService instantiates order matching service
func NewMatcherService(
	och <-chan Order,
	oTimeout int,
	log *logrus.Logger,
) Matcher {
	return &matcherService{
		och:      och,
		oTimeout: oTimeout,
		log:      log,
		buy:      make(map[int][]Order),
		sell:     make(map[int][]Order),
	}
}

// ReceiveOrders gets the orders from a channel and stores in memory
func (m *matcherService) ExecuteOrders() {
	m.log.WithFields(logrus.Fields{
		"OrderTimeout": m.oTimeout}).Info("Starting to Execute Orders")
	for {
		select {
		case o := <-m.och:
			m.log.WithFields(logrus.Fields{
				"OrderId":     o.Id.String(),
				"Transaction": o.Transaction,
				"OrderType":   o.OrderType,
				"Quantity":    o.Quantity,
				"Executed":    o.Executed,
				"Price":       o.Price,
				"OrderTime":   o.OrderTime.Format(time.UnixDate),
			}).Info("Matcher received order")
		}
	}
}
