package matcher

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	id1    = uuid.New()
	id2    = uuid.New()
	id1str = id1.String()
	id2str = id2.String()
	ts1    = time.Now().UTC()
	ts2    = time.Now().UTC()
)

func Test_Matcher_ExecuteOrders(t *testing.T) {
	tests := []struct {
		name       string
		timeout    int
		inOrders   []*Order
		wantOrders map[string]*Order
	}{
		{
			name:    "MatchingPrice, QuantityEqual",
			timeout: 3,
			inOrders: []*Order{
				&Order{
					Id:             id1,
					OrderTime:      ts1,
					Transaction:    Sell,
					PlacedQuantity: 34,
					Quantity:       34,
					Price:          821,
					OrderType:      Market,
				},
				&Order{
					Id:             id2,
					OrderTime:      ts2,
					Transaction:    Buy,
					PlacedQuantity: 34,
					Quantity:       34,
					Price:          821,
					OrderType:      Market,
				},
			},
			wantOrders: map[string]*Order{
				id1.String(): &Order{
					Id:             id1,
					OrderTime:      ts1,
					Transaction:    Sell,
					PlacedQuantity: 34,
					Quantity:       0,
					Executed:       34,
					Price:          821,
					OrderType:      Market,
					Status:         Completed,
				},
				id2.String(): &Order{
					Id:             id2,
					OrderTime:      ts2,
					Transaction:    Buy,
					PlacedQuantity: 34,
					Quantity:       0,
					Executed:       34,
					Price:          821,
					OrderType:      Market,
					Status:         Completed,
				},
			},
		},
		{
			name:    "MatchingPrice, QuantityUnEqual",
			timeout: 3,
			inOrders: []*Order{
				&Order{
					Id:             id1,
					OrderTime:      ts1,
					Transaction:    Sell,
					PlacedQuantity: 34,
					Quantity:       34,
					Price:          821,
					OrderType:      Market,
				},
				&Order{
					Id:             id2,
					OrderTime:      ts2,
					Transaction:    Buy,
					PlacedQuantity: 27,
					Quantity:       27,
					Price:          821,
					OrderType:      Market,
				},
			},
			wantOrders: map[string]*Order{
				id1.String(): &Order{
					Id:             id1,
					OrderTime:      ts1,
					Transaction:    Sell,
					PlacedQuantity: 34,
					Quantity:       7,
					Executed:       27,
					Price:          821,
					OrderType:      Market,
					Status:         TimedOut,
				},
				id2.String(): &Order{
					Id:             id2,
					OrderTime:      ts2,
					Transaction:    Buy,
					PlacedQuantity: 27,
					Quantity:       0,
					Executed:       27,
					Price:          821,
					OrderType:      Market,
					Status:         Completed,
				},
			},
		},
		{
			name:    "NonMatchingPrice",
			timeout: 3,
			inOrders: []*Order{
				&Order{
					Id:             id1,
					OrderTime:      ts1,
					Transaction:    Sell,
					PlacedQuantity: 34,
					Quantity:       34,
					Price:          567,
					OrderType:      Market,
				},
				&Order{
					Id:             id2,
					OrderTime:      ts2,
					Transaction:    Buy,
					PlacedQuantity: 27,
					Quantity:       27,
					Price:          821,
					OrderType:      Market,
				},
			},
			wantOrders: map[string]*Order{
				id1.String(): &Order{
					Id:             id1,
					OrderTime:      ts1,
					Transaction:    Sell,
					PlacedQuantity: 34,
					Quantity:       34,
					Executed:       0,
					Price:          567,
					OrderType:      Market,
					Status:         TimedOut,
				},
				id2.String(): &Order{
					Id:             id2,
					OrderTime:      ts2,
					Transaction:    Buy,
					PlacedQuantity: 27,
					Quantity:       27,
					Executed:       0,
					Price:          821,
					OrderType:      Market,
					Status:         TimedOut,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logrus.New()
			log.SetOutput(ioutil.Discard)

			orders := make(chan *Order)
			complete := make(chan *Order)

			match := NewMatcherService(orders, complete, tt.timeout, log)
			go match.ExecuteOrders()

			for _, o := range tt.inOrders {
				orders <- o
			}

			rmap := make(map[string]*Order)
			for i := 0; i < len(tt.inOrders); i++ {
				o := <-complete
				rmap[o.Id.String()] = o
			}

			assert.Equal(t, tt.wantOrders, rmap)
		})
	}
}
