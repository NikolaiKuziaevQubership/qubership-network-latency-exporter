package collector

import (
	"context"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type contextKey string

const (
	defaultEnabled              = true
	ContextKey       contextKey = "ctxKey"
	commonLabel                 = "dashboard"
	commonLabelValue            = "network-latency-exporter"
)

var (
	factories      = make(map[string]func(logger log.Logger) (Collector, error))
	collectorState = make(map[string]bool)
)

// Collector is minimal interface that let you add new prometheus metrics to network_latency_exporter.
type Collector interface {
	// Name of the Scraper. Should be unique.
	Name() string

	Type() Type

	Initialize(ctx context.Context, config interface{}) error

	// Scrape collects data from database connection and sends it over channel as prometheus metric.
	Scrape(ctx context.Context, metrics *Metrics, ch chan<- prometheus.Metric) error

	Close()
}

func GetCollectorStates() map[string]bool {
	return collectorState
}

func GetCollector(name string, logger log.Logger) (Collector, error) {
	return factories[name](log.With(logger, "depth_caller", log.Caller(4)))
}
