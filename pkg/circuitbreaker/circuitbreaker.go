package circuitbreaker

import (
	"errors"
	"net/http"
	"time"

	"github.com/sony/gobreaker"

	"onmi/pkg/logger"
)

var (
	ErrorCircuitBreakerOpened          = errors.New("circuit breaker opened")
	ErrorCircuitBreakerTooManyRequests = errors.New("circuit breaker too many requests")
	ErrorCircuitBreakerInvalidResponse = errors.New("invalid response type from circuit breaker")
)

type Config struct {
	Interval          time.Duration `yaml:"interval"`
	Timeout           time.Duration `yaml:"timeout"`
	FailureRatio      float64       `yaml:"failureRatio"`
	TotalRequestCount int64         `yaml:"totalRequestCount"`
}

type Proxy struct {
	cb            *gobreaker.CircuitBreaker
	nextTransport http.RoundTripper
	logger        logger.Logger
}

func NewProxy(
	serviceName string,
	cfg *Config,
	log logger.Logger,
	nextTransport http.RoundTripper,
) *Proxy {
	settings := gobreaker.Settings{
		Name:     serviceName,
		Interval: cfg.Interval,
		Timeout:  cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= uint32(cfg.TotalRequestCount) && failureRatio >= cfg.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Логируем при каждой смене state
			log.Debug("state changed", logger.Fields{
				"from": from.String(),
				"to":   to.String(),
				"name": name,
			})
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}
	return &Proxy{
		cb:            gobreaker.NewCircuitBreaker(settings),
		nextTransport: nextTransport,
		logger:        log,
	}
}

func (p *Proxy) RoundTrip(req *http.Request) (*http.Response, error) {
	cbResp, err := p.cb.Execute(func() (interface{}, error) {
		resp, err := p.nextTransport.RoundTrip(req)
		return resp, err
	})

	if errors.Is(err, gobreaker.ErrOpenState) {
		return nil, ErrorCircuitBreakerOpened
	}

	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		return nil, ErrorCircuitBreakerTooManyRequests
	}

	if err != nil {
		return nil, err
	}

	resp, ok := cbResp.(*http.Response)
	if !ok {
		return nil, ErrorCircuitBreakerInvalidResponse
	}

	return resp, nil
}
