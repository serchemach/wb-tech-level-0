package httptransport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	datamodel "github.com/serchemach/wb-tech-level-0/data_model"
)

type Store interface {
	AddOrder(*datamodel.Order) error
	GetOrder(orderUid string) (*datamodel.Order, error)
}

type HTTPTransport interface {
	Listen(context.Context, string)
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

func (t *httpTransport) Listen(ctx context.Context, url string) {
	router := http.NewServeMux()
	router.HandleFunc("GET /api/v1/order", t.OrderHandler)
	router.HandleFunc("GET /", t.InterfaceHandler)

	srv := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     router,
		Addr:        url,
	}
	srv.ListenAndServe()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			t.logger.Error("Error while serving with http transport", "error", err)
			return
		}
		t.logger.Debug("Shut down the http server")
	}()
}

func New(store Store, logger *slog.Logger) HTTPTransport {
	return &httpTransport{
		logger: logger,
		store:  store,
	}
}
