package transport

import (
	"fmt"
	"net/http"

	"github.com/coinbase/cdpcurl/internal/auth"
)

type APIKey = auth.APIKey

var (
	WithENVVariableNames = auth.WithENVVariableNames
	WithENVOnly          = auth.WithENVOnly
	WithFileName         = auth.WithFileName
	WithFileOnly         = auth.WithFileOnly
	WithPath             = auth.WithPath
)

type transport struct {
	originalTransport http.RoundTripper
	authenticator     *auth.Authenticator
	serviceName       string
}

type Option func(o *options)

type options struct {
	apiKey *auth.APIKey

	apiKeyOptions []auth.LoadAPIKeyOption
}

func WithAPIKeyLoaderOption(opt auth.LoadAPIKeyOption) Option {
	return func(o *options) {
		o.apiKeyOptions = append(o.apiKeyOptions, opt)
	}
}

func WithAPIKey(apiKey *auth.APIKey) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

func New(service string, originalTransport http.RoundTripper, opts ...Option) (http.RoundTripper, error) {
	o := &options{
		apiKeyOptions: make([]auth.LoadAPIKeyOption, 0),
	}
	for _, opt := range opts {
		opt(o)
	}
	var authenticator *auth.Authenticator
	if o.apiKey != nil {
		authenticator = auth.NewFromConfig(*o.apiKey)
	} else {
		var err error
		authenticator, err = auth.New(o.apiKeyOptions...)
		if err != nil {
			return nil, err
		}
	}

	return &transport{
		originalTransport: originalTransport,
		authenticator:     authenticator,
		serviceName:       service,
	}, nil
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	jwt, err := t.authenticator.BuildJWT(
		t.serviceName,
		[]string{fmt.Sprintf("%s %s%s", req.Method, req.URL.Host, req.URL.Path)},
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	return t.originalTransport.RoundTrip(req)
}
