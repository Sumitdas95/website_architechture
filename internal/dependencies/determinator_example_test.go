package dependencies

import (
	"fmt"
	"log"
	"net/http"

	"github.com/deliveroo/apm-go"
	"github.com/deliveroo/determinator-go"
	"github.com/deliveroo/test-sonarqube/internal/config"
)

func ExampleInitDeterminator() {
	cfg, _ := config.Load()

	logger, err := newLogger(&cfg)
	if err != nil {
		fmt.Printf("failed to initialize logger: %s\n", err.Error())
		return
	}

	apmService, err := NewAPM(&cfg, apm.WithLogger(logger))
	if err != nil {
		fmt.Printf("failed to initialize APM: %s\n", err.Error())
		return
	}

	circuitManager := newCircuitBreakerManager(cfg)

	httpClientFactory := NewHTTPClientFactory(cfg.Circuit, circuitManager, apmService, http.DefaultClient)

	det, err := InitDeterminator(cfg, httpClientFactory)
	if err != nil {
		log.Fatal(err)
	}
	feature, err := det.Retrieve("example_feature_flag")
	if err != nil {
		fmt.Println(err)
	}
	on, err := feature.IsFeatureFlagOn(determinator.Actor{})
	if err != nil {
		fmt.Println(err)
		return
	}
	if on {
		fmt.Println("The feature flag is on!")
	}
	// Output:
	// determinator: making request: Get "/example_feature_flag": failed to perform roundtrip using circuitbreaker: unsupported protocol scheme ""
}
