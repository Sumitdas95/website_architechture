package httpserver

import (
	"github.com/gorilla/mux"

	"github.com/deliveroo/apm-go/integrations/gorillatrace"
	"github.com/deliveroo/test-sonarqube/internal/dependencies"
	"github.com/deliveroo/test-sonarqube/internal/httpserver/handlers"
)

func NewRouter(deps *dependencies.Dependencies) *mux.Router {
	r := mux.NewRouter()
	orderHandlersHTTPClient, err := deps.HTTPClientFactory.Create("order_handlers", nil)
	if err != nil {
		return nil
	}

	orderHandlers := handlers.OrderHandlers{
		Repository:   deps.Repository,
		Determinator: deps.Determinator,
		Client:       orderHandlersHTTPClient,
	}

	externalHandlers := handlers.NewExternalHandlersFunc(deps.APM, orderHandlersHTTPClient)
	pingHandlers := handlers.Ping{}

	r.HandleFunc("/orders/{id:[0-9]+}", orderHandlers.Get)
	r.HandleFunc("/external", externalHandlers.Get)
	r.HandleFunc("/ping", pingHandlers.Get)

	r.Use(gorillatrace.TracingWithStatusError(deps.APM))

	// Use gorillatrace.SpanLogging(deps.APM) to print a log line for every HTTP request.
	// Be aware that this could be very costly and should not be enabled in production.

	return r
}
