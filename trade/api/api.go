package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/golang/gddo/httputil/header"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/nbasker/tools/trade/matcher"
	"github.com/nbasker/tools/trade/store"
)

// Api REST Service
type Api interface {
	// Run the REST API service
	Run()
}

// apiService defines implementation of the REST Service
type apiService struct {
	endpoint string
	och      chan<- matcher.Order
	retrieve store.Store
	log      *logrus.Logger
}

// NewApiService returns a new apiService
func NewApiService(ep string,
	och chan<- matcher.Order,
	retrieve store.Store,
	log *logrus.Logger,
) Api {
	return &apiService{
		endpoint: ep,
		och:      och,
		retrieve: retrieve,
		log:      log,
	}
}

func (a *apiService) Run() {
	a.log.WithFields(logrus.Fields{
		"endpoint": a.endpoint,
	}).Info("Starting REST Api Service")
	http.HandleFunc("/trade", a.PlaceOrder)
	http.HandleFunc("/orders", a.GetOrders)
	http.ListenAndServe(a.endpoint, nil)
}

func (a *apiService) PlaceOrder(w http.ResponseWriter, req *http.Request) {

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		a.log.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Error("Unable to parse http req")
	}
	a.log.WithFields(logrus.Fields{"req": string(dump)}).Debug("Http Request")

	if req.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(req.Header, "Content-Type")
		if value != "application/json" {
			a.log.Error("Unsupported Content-Type")
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	var order matcher.Order
	if err := json.NewDecoder(req.Body).Decode(&order); err != nil {
		a.log.Error("Unable to decode order")
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	// allocate unique orderId
	order.Id = uuid.New()
	order.OrderTime = time.Now().UTC()
	order.Executed = 0
	resp := fmt.Sprintf("Received Order [%s, %s, %d, %d], Id = %s\n",
		order.Transaction.String(),
		order.OrderType.String(),
		order.Quantity,
		order.Price,
		order.Id.String())

	a.log.WithFields(logrus.Fields{"details": resp}).Debug("Order Received")

	a.och <- order

	io.WriteString(w, resp)
}

func (a *apiService) GetOrders(w http.ResponseWriter, req *http.Request) {
	a.log.WithFields(logrus.Fields{
		"Host":   req.URL.Host,
		"Path":   req.URL.Path,
		"Method": req.Method,
	}).Info("Received")
	a.retrieve.RetrieveExecutedOrders()
	io.WriteString(w, "Printing Orders in log\n")
}
