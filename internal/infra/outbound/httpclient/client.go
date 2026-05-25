package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"market-core/internal/platform/metrics"
)

const (
	defaultMaxRetries     = 3
	defaultOpenTimeout    = 30 * time.Second
	defaultMaxFailures    = 5
	defaultHalfOpenProbes = 2
)

type cbState int

const (
	stateClosed cbState = iota
	stateOpen
	stateHalfOpen
)

type circuitBreaker struct {
	mu           sync.Mutex
	state        cbState
	failures     int
	maxFailures  int
	openTimeout  time.Duration
	openedAt     time.Time
	probeSuccess int
	probeLimit   int
	log          *slog.Logger
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return true
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.openTimeout {
			cb.state = stateHalfOpen
			cb.probeSuccess = 0
			cb.log.Warn("circuit breaker half-open")
			return true
		}
		return false
	case stateHalfOpen:
		return true
	}
	return false
}

func (cb *circuitBreaker) record(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		if success {
			cb.failures = 0
		} else {
			cb.failures++
			if cb.failures >= cb.maxFailures {
				cb.state = stateOpen
				cb.openedAt = time.Now()
				cb.log.Warn("circuit breaker open", "failures", cb.failures)
			}
		}
	case stateHalfOpen:
		if success {
			cb.probeSuccess++
			if cb.probeSuccess >= cb.probeLimit {
				cb.state = stateClosed
				cb.failures = 0
				cb.log.Warn("circuit breaker closed")
			}
		} else {
			cb.state = stateOpen
			cb.openedAt = time.Now()
			cb.log.Warn("circuit breaker open from half-open")
		}
	case stateOpen:
		// no-op: transitions out of Open are handled by allow()
	}
}

type Config struct {
	BaseURL    string
	Target     string
	MaxRetries int
}

type Response struct {
	StatusCode int
	Body       []byte
}

type Client struct {
	cfg     Config
	http    *http.Client
	cb      *circuitBreaker
	metrics *metrics.Metrics
	log     *slog.Logger
	tracer  oteltrace.Tracer
}

func New(cfg Config, m *metrics.Metrics, log *slog.Logger) *Client {
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	return &Client{
		cfg:     cfg,
		http:    &http.Client{Timeout: 30 * time.Second},
		metrics: m,
		log:     log,
		tracer:  otel.Tracer("httpclient"),
		cb: &circuitBreaker{
			maxFailures: defaultMaxFailures,
			openTimeout: defaultOpenTimeout,
			probeLimit:  defaultHalfOpenProbes,
			log:         log,
		},
	}
}

func (c *Client) Do(ctx context.Context, method, path string, body []byte) (*Response, error) {
	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("%s %s", method, path),
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			semconv.HTTPRequestMethodKey.String(method),
			attribute.String("http.url", c.cfg.BaseURL+path),
		),
	)
	defer span.End()

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := jitterBackoff(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		if !c.cb.allow() {
			return nil, fmt.Errorf("circuit breaker open for %s", c.cfg.Target)
		}

		resp, err := c.doOnce(ctx, method, path, body)
		if err != nil {
			c.cb.record(false)
			lastErr = err
			continue
		}

		retryable := resp.StatusCode == 429 ||
			resp.StatusCode == 502 ||
			resp.StatusCode == 503 ||
			resp.StatusCode == 504 ||
			resp.StatusCode >= 500

		c.recordMetric(method, resp.StatusCode)
		span.SetAttributes(semconv.HTTPResponseStatusCodeKey.Int(resp.StatusCode))

		if retryable && attempt < c.cfg.MaxRetries {
			c.cb.record(false)
			lastErr = fmt.Errorf("retryable status %d", resp.StatusCode)
			continue
		}

		c.cb.record(!retryable)
		return resp, nil
	}
	return nil, fmt.Errorf("all %d attempts failed for %s: %w", c.cfg.MaxRetries+1, c.cfg.Target, lastErr)
}

func (c *Client) doOnce(ctx context.Context, method, path string, body []byte) (*Response, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
}

func (c *Client) recordMetric(method string, status int) {
	if c.metrics == nil {
		return
	}
	c.metrics.OutboundTotal.WithLabelValues(c.cfg.Target, method, strconv.Itoa(status)).Inc()
}

func jitterBackoff(attempt int) time.Duration {
	cap := 30 * time.Second
	base := 100 * time.Millisecond
	exp := time.Duration(math.Pow(2, float64(attempt))) * base
	if exp > cap {
		exp = cap
	}
	return time.Duration(rand.Int64N(int64(exp))) //nolint:gosec
}
