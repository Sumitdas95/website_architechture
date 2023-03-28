package dependencies

import (
	"github.com/deliveroo/bnt-internal-test-go/internal/config"
	"github.com/deliveroo/determinator-go"
)

func InitDeterminator(cfg config.Config, httpClientFactory HTTPClientFactory) (*determinator.CachedRetriever, error) {
	httpClient, err := httpClientFactory.Create("determinator", nil)
	if err != nil {
		return nil, err
	}

	r := determinator.NewHTTPRetriever(cfg.Determinator.URL, httpClient, determinator.WithBasicAuthCredentials(cfg.Determinator.Username, cfg.Determinator.Password), determinator.WithUserAgent(cfg.Determinator.UserAgent))
	return determinator.NewCachedRetriever(&r, cfg.Determinator.CacheTTL), nil
}
