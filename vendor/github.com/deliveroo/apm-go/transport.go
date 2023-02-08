package apm

type transportConfig struct {
	disableServiceTag bool
}

func parseTransportOptions(options []TransportOption) *transportConfig {
	opts := &transportConfig{}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}

// TransportOption is a configuration option for transports.
type TransportOption func(*transportConfig)

// WithTransportServiceTagDisabled removes service tag from outgoing transport spans.
func WithTransportServiceTagDisabled() TransportOption {
	return func(opts *transportConfig) {
		opts.disableServiceTag = true
	}
}
