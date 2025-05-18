package main

import (
	"log"
	"os"
	"strconv"

	"github.com/anekazek/simple-blockchain/pkg/api"
	"github.com/anekazek/simple-blockchain/pkg/blockchain"
	"github.com/anekazek/simple-blockchain/pkg/metrics"
)

func main() {
	// Set mining difficulty (can be made configurable via flags/env)
	difficulty := 1
	if os.Getenv("BLOCKCHAIN_DIFFICULTY") != "" {
		val, err := strconv.Atoi(os.Getenv("BLOCKCHAIN_DIFFICULTY"))
		if err == nil && val > 0 {
			difficulty = val
		}
	}

	// Initialize blockchain with genesis block
	chain := blockchain.NewBlockchain()

	// Initialize transaction pool
	txPoolSize := 1000
	if os.Getenv("TX_POOL_SIZE") != "" {
		val, err := strconv.Atoi(os.Getenv("TX_POOL_SIZE"))
		if err == nil && val > 0 {
			txPoolSize = val
		}
	}
	txPool := blockchain.NewTransactionPool(txPoolSize)

	// Initialize metrics
	blockchainMetrics := metrics.NewBlockchainMetrics()
	metricsPort := "9090"
	if os.Getenv("METRICS_PORT") != "" {
		metricsPort = os.Getenv("METRICS_PORT")
	}
	blockchainMetrics.StartServer(metricsPort)

	// Set initial node health to healthy
	blockchainMetrics.SetNodeHealth(true)

	// Get API ports
	httpPort := "8080"
	if os.Getenv("HTTP_PORT") != "" {
		httpPort = os.Getenv("HTTP_PORT")
	}

	wsPort := "8081"
	if os.Getenv("WS_PORT") != "" {
		wsPort = os.Getenv("WS_PORT")
	}

	// Create enhanced server with WebSocket support
	server := api.NewEnhancedBlockchainServer(chain, txPool, difficulty, blockchainMetrics)

	// Configure TLS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		server.ConfigureTLS(certFile, keyFile)
		log.Println("TLS enabled for API and WebSocket servers")
	}

	log.Printf("Starting blockchain with difficulty: %d\n", difficulty)
	log.Printf("Transaction pool initialized with capacity: %d\n", txPoolSize)
	log.Printf("Metrics server available at http://localhost:%s/metrics\n", metricsPort)
	log.Printf("Web dashboard available at http://localhost:%s\n", httpPort)

	// Start the enhanced server
	log.Fatal(server.Start(httpPort, wsPort))
}
