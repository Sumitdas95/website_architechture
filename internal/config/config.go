package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	envPrefix = "settings"
)

// Server contains the configuration for the HTTP server.
type Server struct {
	// IdleTimeout is the maximum amount of time to wait for an open connection
	// when processing no requests and keep-alives are enabled. If this value is
	// 0, ReadTimeout value be used.
	IdleTimeout time.Duration `envconfig:"HTTP_SERVER_IDLE_TIMEOUT" default:"60s"`

	// Port is the HTTP server port.
	Port int `envconfig:"PORT" default:"3000"`

	// ReadTimeout is the maximum duration for reading the entire request,
	// including the body.
	ReadTimeout time.Duration `envconfig:"HTTP_SERVER_READ_TIMEOUT" default:"1s"`

	// WriteTimeout is the maximum duration before timing out
	// writes of the response.
	WriteTimeout time.Duration `envconfig:"HTTP_SERVER_WRITE_TIMEOUT" default:"2s"`
}

// Settings holds application-specific config.
type Settings struct {
	SpanLogging     bool `envconfig:"SPAN_LOGGING" envDefault:"false"`   // Write spans to logger, for debug purpose
	StatsDLogging   bool `envconfig:"STATSD_LOGGING" envDefault:"false"` // Write statsd events to logger
	SuppressLogging bool `envconfig:"SUPPRESS_LOGGING" default:"false"`  // Replaces the logger with a Noop
}

// Hopper contains parameters injected from Hopper.
type Hopper struct {
	AppName     string `envconfig:"HOPPER_APP_NAME" default:"bnt-internal-test-go"`
	Environment string `envconfig:"HOPPER_ENVIRONMENT" default:"development"`
	ReleaseID   string `envconfig:"HOPPER_RELEASE_ID"`
	ServiceName string `envconfig:"HOPPER_SERVICE_NAME"`
}

// Database contains configuration for the Postgres Database.
type Database struct {
	URL       string `envconfig:"DATABASE_URL" default:"postgres://localhost:5434/service_template_go_development?sslmode=disable"`
	ReaderURL string `envconfig:"DATABASE_URL_READER" default:"postgres://localhost:5434/service_template_go_development?sslmode=disable"`
}

// Datadog contains configuration for the Datadog APM.
type Datadog struct {
	AppName     string `envconfig:"HOPPER_APP_NAME" default:"bnt-internal-test-go"`
	ServiceName string `envconfig:"HOPPER_SERVICE_NAME" default:"app"`
	Env         string `envconfig:"STATSD_ENV" default:"development"`
	Host        string `envconfig:"STATSD_HOST"`
	StatsDPort  uint   `envconfig:"STATSD_PORT"`
	TracerPort  uint   `envconfig:"DATADOG_TRACER_PORT"`
}

type Determinator struct {
	URL       *url.URL      `envconfig:"DETERMINATOR_URL"`
	Username  string        `envconfig:"DETERMINATOR_USERNAME"`
	Password  string        `envconfig:"DETERMINATOR_PASSWORD"`
	CacheTTL  time.Duration `envconfig:"DETERMINATOR_CACHE_TTL"`
	UserAgent string        `envconfig:"DETERMINATOR_USER_AGENT"`
}

// Circuit contains configuration for HTTP circuit breaking.
// Missing options use the defaults provided by the hystrix package.
// Applications may want to create separate configuration for different HTTP
// clients, to allow per-service circuit breaking configuration.
type Circuit struct {
	Timeout                int `envconfig:"HTTP_CIRCUIT_TIMEOUT"`
	MaxConcurrentRequests  int `envconfig:"HTTP_CIRCUIT_MAX_CONCURRENT_REQUESTS"`
	RequestVolumeThreshold int `envconfig:"HTTP_CIRCUIT_REQUEST_VOLUME_THRESHOLD"`
	SleepWindow            int `envconfig:"HTTP_CIRCUIT_SLEEP_WINDOW"`
	ErrorPercentThreshold  int `envconfig:"HTTP_CIRCUIT_ERROR_PERCENT_THRESHOLD"`
}

// Sentry contains configuration for Sentry's error reporting.
type Sentry struct {
	DSN     string `envconfig:"SENTRY_DSN"`
	Env     string `envconfig:"SENTRY_ENVIRONMENT"`
	Release string `envconfig:"SENTRY_RELEASE"`
}

// Config is the global config struct.
type Config struct {
	Env          string // APP_ENV
	Settings     Settings
	Server       Server
	Hopper       Hopper
	Database     Database
	Datadog      Datadog
	Circuit      Circuit
	Determinator Determinator
	Sentry       Sentry
}

// Load configuration from environment.
func Load() (Config, error) {
	config := Config{}

	if err := envconfig.Process(envPrefix, &config); err != nil {
		return config, fmt.Errorf("failed to load config from environment: %w", err)
	}

	return config, nil
}
