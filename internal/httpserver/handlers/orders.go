package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/deliveroo/bnt-internal-test-go/internal/httpserver/gorillautils"
	"github.com/deliveroo/bnt-internal-test-go/internal/orders"
	"github.com/deliveroo/determinator-go"
)

type Order struct {
	ID     int
	Status string
}

type OrderHandlers struct {
	Repository   orders.Repository
	Determinator determinator.Retriever
	Client       *http.Client
}

func (o *OrderHandlers) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	temp := Order{orderID, "FULFILLED"}

	_ = gorillautils.RenderJSON(w, temp)
}
