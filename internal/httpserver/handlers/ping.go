package handlers

import (
	"net/http"
)

type Ping struct{}

func (o *Ping) Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
