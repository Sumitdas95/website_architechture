package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/deliveroo/determinator-go"
	"github.com/deliveroo/test-sonarqube/internal/httpserver/gorillautils"
	"github.com/deliveroo/test-sonarqube/internal/orders"
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

	context := r.Context()
	order, err := o.Repository.GetOrder(context, orderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if order == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = gorillautils.RenderJSON(w, order)
}
