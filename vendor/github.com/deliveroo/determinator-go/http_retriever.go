package determinator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Doer is a generic HTTP client interface.
// It is implemented by net/http's *http.Client, as well as other more advanced
// clients like heimdall.Client.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// HTTPRetriever loads features from JSON encoded GET endpoints.
type HTTPRetriever struct {
	host        *url.URL
	http        Doer
	credentials *Credentials
	userAgent   string
}

// Opt is an option passed to HTTPRetriever.
type Opt func(*HTTPRetriever)

// Credentials stores the username and password.
type Credentials struct {
	username string
	password string
}

// WithBasicAuthCredentials allows setting BasicAuth Credentials on Doer,
// which implements *http.Client.
func WithBasicAuthCredentials(username, password string) Opt {
	return func(r *HTTPRetriever) {
		r.credentials = &Credentials{
			username: username,
			password: password,
		}
	}
}

const userAgentHeader = "User-Agent"

// WithUserAgent allows setting User Agent on Doer, which implements
// *http.Client.
func WithUserAgent(ua string) Opt {
	return func(r *HTTPRetriever) {
		r.userAgent = ua
	}
}

// NewHTTPRetriever creates a new NewHTTPRetriever object.
func NewHTTPRetriever(host *url.URL, http Doer, opts ...Opt) HTTPRetriever {
	retriever := &HTTPRetriever{
		host: host,
		http: http,
	}

	for _, o := range opts {
		o(retriever)
	}

	return *retriever
}

// Retrieve loads feature via HTTP by feature identifier.
func (retriever *HTTPRetriever) Retrieve(featureID string) (Feature, error) {
	featureIDPath, err := url.Parse(url.PathEscape(featureID))
	if err != nil {
		return nil, err
	}
	if retriever.host == nil {
		return nil, errors.New("host URL cannot be nil")
	}
	featureURL := retriever.host.ResolveReference(featureIDPath)

	req, err := http.NewRequest(http.MethodGet, featureURL.String(), bytes.NewBuffer(nil))
	if err != nil {
		return nil, fmt.Errorf("determinator: creating request: %w", err)
	}

	if retriever.credentials != nil {
		req.SetBasicAuth(retriever.credentials.username, retriever.credentials.password)
	}
	if retriever.userAgent != "" {
		req.Header.Add(userAgentHeader, retriever.userAgent)
	}

	response, err := retriever.http.Do(req)
	defer func() {
		if response != nil {
			if closeError := response.Body.Close(); closeError != nil && err == nil {
				err = errors.Wrap(closeError, "determinator: failed to close response body")
			}
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("determinator: making request: %w", err)
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("determinator: resource does not exist %s", response.Status)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("determinator: received status %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("determinator: reading response body: %w", err)
	}

	feature, err := ParseJSONFeature(body, true)
	if err != nil {
		return nil, fmt.Errorf("determinator: parsing feature JSON: %w", err)
	}

	err = feature.validate()
	if err != nil {
		return nil, fmt.Errorf("determinator: validating feature: %w", err)
	}
	return feature, nil
}
