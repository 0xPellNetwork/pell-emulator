package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/0xPellNetwork/pell-emulator/libs/log"
	"github.com/0xPellNetwork/pell-emulator/libs/utils"
)

// PellMetrics contains instrumented metrics that should be incremented by the avs node using the methods below
type PellMetrics struct {
	ipPortAddress string
	logger        log.Logger
	// metrics
	// fees are not yet turned on, so these should just be 0 for the time being
	feeEarnedTotal   *prometheus.CounterVec
	performanceScore prometheus.Gauge
}

var _ Metrics = (*PellMetrics)(nil)

// Follows the structure from https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#hdr-A_Basic_Example
// TODO(samlaf): I think each avs runs in a separate docker bridge network.
// In order for prometheus to scrape the metrics does the address need to be 0.0.0.0:port to accept connections from other networks?
func NewPellMetrics(avsName, ipPortAddress string, reg prometheus.Registerer, logger log.Logger) *PellMetrics {

	metrics := &PellMetrics{
		feeEarnedTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   "bin",
				Name:        "fees_earned_total",
				Help:        "The amount of fees earned in <token>",
				ConstLabels: prometheus.Labels{"avs_name": avsName},
			},
			[]string{"token"},
		),
		performanceScore: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "bin",
				Name:        "performance_score",
				Help:        "The performance metric is a score between 0 and 100 and each developer can define their own way of calculating the score. The score is calculated based on the performance of the Node and the performance of the backing services.",
				ConstLabels: prometheus.Labels{"avs_name": avsName},
			},
		),
		ipPortAddress: ipPortAddress,
		logger:        logger,
	}

	metrics.initMetrics()
	return metrics
}

func (m *PellMetrics) initMetrics() {
	// Performance score starts as 100, and goes down if node doesn't perform well
	m.performanceScore.Set(100)

	// TODO(samlaf): should we initialize the feeEarnedTotal? This would require the user to pass in a list of tokens for which to initialize the metric
	// same for rpcRequestDurationSeconds and rpcRequestTotal... we could initialize them to be 0 on every json-rpc... but is that really necessary?
}

// AddPellFeeEarnedTotal adds the fee earned to the total fee earned metric
func (m *PellMetrics) AddFeeEarnedTotal(amount float64, token string) {
	m.feeEarnedTotal.WithLabelValues(token).Add(amount)
}

// SetPerformanceScore sets the performance score of the node
func (m *PellMetrics) SetPerformanceScore(score float64) {
	m.performanceScore.Set(score)
}

// Start creates an http handler for reg and starts the prometheus server in a goroutine, listening at m.ipPortAddress.
// reg needs to be the prometheus registry that was passed in the NewPellMetrics constructor
func (m *PellMetrics) Start(ctx context.Context, reg prometheus.Gatherer) <-chan error {
	m.logger.Info("Starting metrics server at port", "port", m.ipPortAddress)
	errChan := make(chan error, 1)
	mux := http.NewServeMux()
	httpServer := http.Server{
		Addr:    m.ipPortAddress,
		Handler: mux,
	}
	mux.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))

	// shutdown server on context done
	go func() {
		<-ctx.Done()
		m.logger.Info("shutdown signal received")
		defer func() {
			close(errChan)
		}()

		if err := httpServer.Shutdown(context.Background()); err != nil {
			errChan <- err
		}
		m.logger.Info("shutdown completed")
	}()

	go func() {
		err := httpServer.ListenAndServe()
		if err == http.ErrServerClosed {
			m.logger.Info("server closed")
		} else {
			errChan <- utils.WrapError("Prometheus server failed", err)
		}
	}()
	return errChan
}
