package main

import (
	"log"
	"os"
	"strconv"

	"github.com/anekazek/simple-blockchain/pkg/api"
	"github.com/anekazek/simple-blockchain/pkg/blockchain"
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

	// Create and start HTTP server
	server := api.NewBlockchainServer(chain, difficulty)
	port := "8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	log.Printf("Starting blockchain with difficulty: %d\n", difficulty)
	log.Fatal(server.Start(port))
}
