package metrics

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// BlockchainMetrics collects and exposes blockchain metrics
type BlockchainMetrics struct {
	// Metrics collectors
	blockCounter       prometheus.Counter
	blockTime          prometheus.Histogram
	transactionCounter prometheus.Counter
	transactionTime    prometheus.Histogram
	peerCount          prometheus.Gauge
	nodeHealth         prometheus.Gauge
	blockSize          prometheus.Histogram
	consensusRoundTime prometheus.Histogram

	// Start time for calculating uptime
	startTime time.Time
}

// NewBlockchainMetrics creates and registers blockchain metrics
func NewBlockchainMetrics() *BlockchainMetrics {
	m := &BlockchainMetrics{
		startTime: time.Now(),
		blockCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "blockchain_blocks_total",
			Help: "The total number of blocks in the blockchain",
		}),
		blockTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "blockchain_block_processing_time_seconds",
			Help:    "Time taken to process and add a new block",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		}),
		transactionCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "blockchain_transactions_total",
			Help: "The total number of transactions processed",
		}),
		transactionTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "blockchain_transaction_processing_time_seconds",
			Help:    "Time taken to process a transaction",
			Buckets: prometheus.LinearBuckets(0.01, 0.01, 10),
		}),
		peerCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_peer_count",
			Help: "The current number of connected peers",
		}),
		nodeHealth: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "blockchain_node_health",
			Help: "Node health status (1 = healthy, 0 = unhealthy)",
		}),
		blockSize: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "blockchain_block_size_bytes",
			Help:    "Size of blocks in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10),
		}),
		consensusRoundTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "blockchain_consensus_round_time_seconds",
			Help:    "Time taken to complete a consensus round",
			Buckets: prometheus.LinearBuckets(0.5, 0.5, 10),
		}),
	}

	// Set initial health to healthy
	m.nodeHealth.Set(1)

	return m
}

// StartServer starts the metrics HTTP server
func (m *BlockchainMetrics) StartServer(port string) {
	// Register the metrics handler
	http.Handle("/metrics", promhttp.Handler())

	// Start the HTTP server in a goroutine
	go func() {
		log.Printf("Metrics server listening on :%s/metrics\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("Metrics server error: %v\n", err)
		}
	}()
}

// BlockAdded records metrics when a new block is added
func (m *BlockchainMetrics) BlockAdded(processingTime time.Duration, blockSizeBytes int) {
	m.blockCounter.Inc()
	m.blockTime.Observe(processingTime.Seconds())
	m.blockSize.Observe(float64(blockSizeBytes))
}

// TransactionProcessed records metrics when a transaction is processed
func (m *BlockchainMetrics) TransactionProcessed(processingTime time.Duration) {
	m.transactionCounter.Inc()
	m.transactionTime.Observe(processingTime.Seconds())
}

// UpdatePeerCount updates the peer count metric
func (m *BlockchainMetrics) UpdatePeerCount(count int) {
	m.peerCount.Set(float64(count))
}

// SetNodeHealth updates the node health status
func (m *BlockchainMetrics) SetNodeHealth(healthy bool) {
	if healthy {
		m.nodeHealth.Set(1)
	} else {
		m.nodeHealth.Set(0)
	}
}

// RecordConsensusRound records the time taken for a consensus round
func (m *BlockchainMetrics) RecordConsensusRound(duration time.Duration) {
	m.consensusRoundTime.Observe(duration.Seconds())
}

// GetUptime returns the node uptime in seconds
func (m *BlockchainMetrics) GetUptime() float64 {
	return time.Since(m.startTime).Seconds()
}
