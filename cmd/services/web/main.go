package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/deliveroo/bnt-internal-test-go/internal/config"
	"github.com/deliveroo/bnt-internal-test-go/internal/dependencies"
	"github.com/deliveroo/bnt-internal-test-go/internal/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load configuration: %s", err)
	}

	deps, err := dependencies.Initialize(cfg)
	if err != nil {
		log.Fatalf("could not load dependencies: %s", err)
	}

	ctx := context.Background()
	log := deps.APM.Logger()
	log.Info("Server has booted!", zap.Int("port", cfg.Server.Port))

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      httpserver.NewRouter(deps),
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	shutdownCompleteChan := handleShutdownSignal(func() {
		// When we call this, ListenAndServe will immediately return
		// http.ErrServerClosed
		if err := server.Shutdown(ctx); err != nil {
			log.Error("server.Shutdown failed", zap.Error(err))
		}
	})

	if err = server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		// Shutdown has been called, we must wait here until it completes
		<-shutdownCompleteChan
	} else {
		log.Error("http.ListenAndServer failed", zap.Error(err))
	}

	deps.Shutdown()
	log.Info("Shutdown gracefully")
}

// handleShutdownSignal awaits SIGINT or SIGTERM then calls onSignalReceived
// asynchronously. It returns a channel that will be closed when shutdown has
// completed.
func handleShutdownSignal(onSignalReceived func()) <-chan struct{} {
	shutdown := make(chan struct{})

	go func() {
		sigNotifier := make(chan os.Signal, 1)
		signal.Notify(sigNotifier, os.Interrupt, syscall.SIGTERM)

		// Park here until a signal is received
		<-sigNotifier

		onSignalReceived()
		close(shutdown)
	}()

	return shutdown
}
