package handlers

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/deliveroo/apm-go"
)

type ExternalHandlers struct {
	APM    apm.Service
	client *http.Client
}

func NewExternalHandlersFunc(a apm.Service, client *http.Client) ExternalHandlers {
	return ExternalHandlers{APM: a, client: client}
}

func (o *ExternalHandlers) Get(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", strings.NewReader(""))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apm.LoggerFromContext(r.Context(), o.APM).Error("failed to create external request", zap.Error(err))
		return
	}

	res, err := o.client.Do(req.WithContext(r.Context()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apm.LoggerFromContext(r.Context(), o.APM).Error("failed to perform external request", zap.Error(err))
		return
	}
	defer res.Body.Close()
	w.WriteHeader(res.StatusCode)
	if _, err = io.Copy(w, res.Body); err != nil {
		apm.LoggerFromContext(r.Context(), o.APM).Error("failed to copy external response", zap.Error(err))
	}
}
