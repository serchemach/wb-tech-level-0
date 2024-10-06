package httptransport

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	datamodel "github.com/serchemach/wb-tech-level-0/data_model"
)

type Store interface {
	AddOrder(*datamodel.Order) error
	GetOrder(orderUid string) (*datamodel.Order, error)
}

type HTTPTransport interface {
	Listen(string) error
	OrderHandler(w http.ResponseWriter, r *http.Request)
	InterfaceHandler(w http.ResponseWriter, r *http.Request)
}

type httpTransport struct {
	store  Store
	logger *slog.Logger
}

func (t *httpTransport) OrderHandler(w http.ResponseWriter, r *http.Request) {
	orderUid := r.URL.Query().Get("order_uid")
	if orderUid == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No order id given"))
		return
	}

	order, err := t.store.GetOrder(orderUid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while fetching the order data: %s", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		t.logger.Error("Error while encoding the json", "error", err)
	}
}

func (t *httpTransport) InterfaceHandler(w http.ResponseWriter, r *http.Request) {
	htmlFile, err := os.ReadFile("pages/form.html")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		t.logger.Error("Error while reading the page file", "error", err)
	}

	w.Write(htmlFile)
}

func (t *httpTransport) Listen(url string) error {
	router := http.NewServeMux()
	router.HandleFunc("GET /api/v1/order", t.OrderHandler)
	router.HandleFunc("GET /", t.InterfaceHandler)

	return http.ListenAndServe(url, router)
}

func New(store Store, logger *slog.Logger) HTTPTransport {
	return &httpTransport{
		logger: logger,
		store:  store,
	}
}
