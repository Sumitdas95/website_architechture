// Package dependencies contains all the application-wide dependencies (such as
// databases, redis connections, config, ...), and functions for conveniently
// loading them on start-up.
//
// Ideally only your command line wrapper (in cmd/services/) will depend on this
// package. The rest of your application (such as groups of HTTP handlers)
// should accept only the dependencies that they require, and you should inject
// these on start-up (such as when configuring HTTP routes).
//
// This package also exposes public functions to expose set-up for individual
// dependencies (e.g. InitDatabase()), which can be called from light-weight
// command line utilities which do not need all of your application dependencies.
package dependencies

import (
	"fmt"
	"net/http"

	"github.com/cep21/circuit/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/deliveroo/apm-go"
	"github.com/deliveroo/determinator-go"
	"github.com/deliveroo/test-sonarqube/internal/config"
	"github.com/deliveroo/test-sonarqube/internal/orders"
)

// Dependencies groups all application dependencies together, for easy start-up.
type Dependencies struct {
	CircuitManager *circuit.Manager
	Config         config.Config

	WriterDB          *pgxpool.Pool
	ReaderDB          *pgxpool.Pool
	Determinator      determinator.Retriever
	HTTPClientFactory HTTPClientFactory
	Repository        orders.Repository
	APM               apm.Service
}

// Initialize loads all application dependencies.
func Initialize(cfg config.Config) (*Dependencies, error) {
	logger, err := newLogger(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	apmService, err := NewAPM(&cfg, apm.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize APM: %w", err)
	}

	writeDB, err := InitDatabase(cfg.Database.URL, apmService)
	if err != nil {
		return nil, err
	}

	readDB, err := InitDatabase(cfg.Database.ReaderURL, apmService)
	if err != nil {
		return nil, err
	}

	circuitManager := newCircuitBreakerManager(cfg)

	httpClientFactory := NewHTTPClientFactory(cfg.Circuit, circuitManager, apmService, http.DefaultClient)

	determinator, err := InitDeterminator(cfg, httpClientFactory)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Determinator: %w", err)
	}

	dependencies := &Dependencies{
		CircuitManager:    circuitManager,
		Config:            cfg,
		WriterDB:          writeDB,
		ReaderDB:          readDB,
		Determinator:      determinator,
		HTTPClientFactory: httpClientFactory,
		Repository:        orders.NewRepository(writeDB, readDB),
		APM:               apmService,
	}

	return dependencies, nil
}

func newLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	switch cfg.Hopper.Environment {
	case "development", "setup":
		if cfg.Settings.SuppressLogging {
			logger = zap.NewNop()
		} else {
			logger, err = zap.NewDevelopment()
		}
	default:
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialise logger: %w", err)
	}

	return logger, nil
}

// Shutdown should be called on application shutdown to allow dependencies to
// shutdown gracefully.
func (d *Dependencies) Shutdown() {
	d.APM.Close()
	CloseDatabaseConnection(d.WriterDB)
}
