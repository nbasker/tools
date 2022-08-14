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

// Status enum
type Status int

const (
	Placed Status = iota + 1
	TimedOut
	Completed
)

func (s Status) String() string {
	switch s {
	case Placed:
		return "placed"
	case TimedOut:
		return "timedout"
	case Completed:
		return "completed"
	}
	return "unknown"
}

// Order defines the order placed for trade
type Order struct {
	Id             uuid.UUID   `json:"id,omitempty"`
	OrderTime      time.Time   `json:"order_time,omitempty"`
	Transaction    Transaction `json:"transaction,omitempty"`
	PlacedQuantity int         `json:"placed_quantity,omitempty"`
	Quantity       int         `json:"quantity,omitempty"`
	Executed       int         `json:"executed,omitempty"`
	Price          int         `json:"price,omitempty"`
	OrderType      OrderType   `json:"order_type,omitempty"`
	Status         Status      `json:"status,omitempty"`
}

type OrderMap map[int][]*Order

// Matcher that receives orders and executes
type Matcher interface {
	// Execute Orders matches the buy and sell order from in memory maps.
	ExecuteOrders()
}

// matcherService implements the order processing
type matcherService struct {
	och      <-chan *Order
	complete chan<- *Order
	oTimeout int
	log      *logrus.Logger
	buy      OrderMap
	sell     OrderMap
}

// NewMatcherService instantiates order matching service
func NewMatcherService(
	och <-chan *Order,
	complete chan<- *Order,
	oTimeout int,
	log *logrus.Logger,
) Matcher {
	return &matcherService{
		och:      och,
		complete: complete,
		oTimeout: oTimeout,
		log:      log,
		buy:      make(OrderMap),
		sell:     make(OrderMap),
	}
}

// ReceiveOrders gets the orders from a channel and stores in memory
func (m *matcherService) ExecuteOrders() {
	m.log.WithFields(logrus.Fields{
		"OrderTimeout": m.oTimeout}).Info("Starting to Execute Orders")
	for {
		select {
		case o := <-m.och:
			m.processOrder(o)
		case <-time.After(15 * time.Second):
			m.log.Info("Clean Timedout Orders")
			// m.printLiveOrders()
			m.cleanTimedoutOrders(Buy)
			m.cleanTimedoutOrders(Sell)
		}
	}
}

func updateOrderQuantity(executed int, in, match *Order) {
	in.Quantity -= executed
	in.Executed += executed
	match.Quantity -= executed
	match.Executed += executed
}

func (m *matcherService) processInputAgainstMatch(in *Order, mlist []*Order) {
	for _, mo := range mlist {
		if in.Quantity > mo.Quantity {
			updateOrderQuantity(mo.Quantity, in, mo)
		} else {
			updateOrderQuantity(in.Quantity, in, mo)
			break
		}
	}

	// check input order is fully executed
	if in.Quantity > 0 {
		// Not fully executed as in.Quantity is not 0
		if in.Transaction == Buy {
			m.buy[in.Price] = append(m.buy[in.Price], in)
		} else if in.Transaction == Sell {
			m.sell[in.Price] = append(m.sell[in.Price], in)
		}
	} else {
		// Fuly executed send order in complete channel
		in.Status = Completed
		m.complete <- in
	}
}

func (m *matcherService) cleanTimedoutOrders(oType Transaction) {
	var oMap OrderMap
	if oType == Buy {
		oMap = m.buy
	} else if oType == Sell {
		oMap = m.sell
	}
	newMap := make(OrderMap)

	for p, ol := range oMap {
		temp := ol[:0]
		for _, o := range ol {
			if int(time.Since(o.OrderTime).Seconds()) > m.oTimeout {
				m.log.WithFields(logrus.Fields{
					"Id":             o.Id.String()[:10],
					"Transaction":    o.Transaction,
					"PlacedQuantity": o.PlacedQuantity,
					"Quantity":       o.Quantity,
					"Executed":       o.Executed,
					"OrderTime":      o.OrderTime.Format(time.UnixDate),
				}).Debug("TimedOut Order")

				o.Status = TimedOut
				m.complete <- o
			} else {
				temp = append(temp, o)
			}
		}
		if len(temp) > 0 {
			newMap[p] = temp
		}
	}

	// Assign the new order-list for each price in the original map
	if oType == Buy {
		m.buy = newMap
	} else if oType == Sell {
		m.sell = newMap
	}
}

func (m *matcherService) cleanCompletedOrders(oType Transaction, price int) {

	var mlist []*Order
	if oType == Buy {
		mlist = m.buy[price]
	} else if oType == Sell {
		mlist = m.sell[price]
	}

	// check in match-list if any order is completed
	// Remove from the copied list and send message in complete channel
	temp := mlist[:0]
	for _, o := range mlist {
		if o.Quantity > 0 {
			temp = append(temp, o)
		} else {
			// Fuly executed send order in complete channel
			o.Status = Completed
			m.complete <- o
		}
	}

	// Reassign to the map
	if len(temp) > 0 && len(temp) != len(mlist) {
		if oType == Buy {
			m.buy[price] = temp
		} else if oType == Sell {
			m.sell[price] = temp
		}
	} else if len(temp) == 0 {
		if oType == Buy {
			delete(m.buy, price)
		} else if oType == Sell {
			delete(m.sell, price)
		}
	}
}

func (m *matcherService) processOrder(o *Order) {
	m.log.WithFields(logrus.Fields{
		"OrderId":     o.Id.String()[:10],
		"Transaction": o.Transaction,
		"OrderType":   o.OrderType,
		"Quantity":    o.Quantity,
		"Executed":    o.Executed,
		"Price":       o.Price,
		"OrderTime":   o.OrderTime.Format(time.UnixDate),
	}).Debug("Matcher received order")

	switch o.Transaction {
	case Buy:
		// Check for sellOrder with a matching price
		sellOrder, ok := m.sell[o.Price]
		if ok {
			m.processInputAgainstMatch(o, sellOrder)
			m.cleanCompletedOrders(Sell, o.Price)
		} else {
			// No match add to buy map
			m.buy[o.Price] = append(m.buy[o.Price], o)
		}
	case Sell:
		// Check for buyOrder with a matching price
		buyOrder, ok := m.buy[o.Price]
		if ok {
			m.processInputAgainstMatch(o, buyOrder)
			m.cleanCompletedOrders(Buy, o.Price)
		} else {
			// No match add to sell map
			m.sell[o.Price] = append(m.sell[o.Price], o)
		}
	default:
		m.log.Error("Invalid order transaction, only Buy or Sell supported")
	}
}

func (m *matcherService) printLiveOrders() {
	m.log.Info("+++Live Orders+++")
	for _, ol := range m.buy {
		for _, o := range ol {
			m.log.WithFields(logrus.Fields{
				"Quantity": o.Quantity,
				"Executed": o.Executed,
				"Price":    o.Price,
				"Id":       o.Id.String()[:10],
			}).Info("BuyOrder")
		}
	}
	for _, ol := range m.sell {
		for _, o := range ol {
			m.log.WithFields(logrus.Fields{
				"Quantity": o.Quantity,
				"Executed": o.Executed,
				"Price":    o.Price,
				"Id":       o.Id.String()[:10],
			}).Info("SellOrder")
		}
	}
}
