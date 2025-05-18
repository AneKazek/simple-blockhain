package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/anekazek/simple-blockchain/pkg/blockchain"
	"github.com/anekazek/simple-blockchain/pkg/contracts"
	"github.com/anekazek/simple-blockchain/pkg/metrics"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// EnhancedBlockchainServer provides a full-featured API with WebSocket support and TLS
type EnhancedBlockchainServer struct {
	chain        *blockchain.Chain
	txPool       *blockchain.TransactionPool
	difficulty   int
	wasmEngine   *contracts.WASMEngine
	luaEngine    *contracts.LuaEngine
	metrics      *metrics.BlockchainMetrics
	clients      map[*websocket.Conn]bool
	broadcast    chan interface{}
	clientsMutex sync.Mutex
	upgrader     websocket.Upgrader
	tlsCertFile  string
	tlsKeyFile   string
	enableTLS    bool
}

// NewEnhancedBlockchainServer creates a new enhanced server
func NewEnhancedBlockchainServer(chain *blockchain.Chain, txPool *blockchain.TransactionPool, difficulty int, metrics *metrics.BlockchainMetrics) *EnhancedBlockchainServer {
	return &EnhancedBlockchainServer{
		chain:      chain,
		txPool:     txPool,
		difficulty: difficulty,
		wasmEngine: contracts.NewWASMEngine(),
		luaEngine:  contracts.NewLuaEngine(),
		metrics:    metrics,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}, 100),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		enableTLS: false,
	}
}

// ConfigureTLS sets up TLS for secure connections
func (s *EnhancedBlockchainServer) ConfigureTLS(certFile, keyFile string) {
	s.tlsCertFile = certFile
	s.tlsKeyFile = keyFile
	s.enableTLS = true
}

// Start initializes the HTTP server with all routes
func (s *EnhancedBlockchainServer) Start(httpPort, wsPort string) error {
	// Start WebSocket server in a separate goroutine
	go s.startWebSocketServer(wsPort)

	// Start broadcasting service
	go s.handleBroadcasts()

	// Create router with all API endpoints
	r := mux.NewRouter()

	// Blockchain endpoints
	r.HandleFunc("/api/blockchain", s.handleGetBlockchain).Methods("GET")
	r.HandleFunc("/api/blocks", s.handleGetBlocks).Methods("GET")
	r.HandleFunc("/api/blocks/{hash}", s.handleGetBlock).Methods("GET")

	// Transaction endpoints
	r.HandleFunc("/api/transactions", s.handleCreateTransaction).Methods("POST")
	r.HandleFunc("/api/transactions", s.handleGetTransactions).Methods("GET")
	r.HandleFunc("/api/transactions/{id}", s.handleGetTransaction).Methods("GET")
	r.HandleFunc("/api/transactions/pending", s.handleGetPendingTransactions).Methods("GET")

	// Smart contract endpoints
	r.HandleFunc("/api/contracts", s.handleDeployContract).Methods("POST")
	r.HandleFunc("/api/contracts", s.handleGetContracts).Methods("GET")
	r.HandleFunc("/api/contracts/{id}", s.handleGetContract).Methods("GET")
	r.HandleFunc("/api/contracts/{id}/execute", s.handleExecuteContract).Methods("POST")

	// Serve static files for the dashboard
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	// Start HTTP server
	log.Printf("API server listening on port %s\n", httpPort)

	if s.enableTLS {
		// Configure TLS
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
		}

		server := &http.Server{
			Addr:      ":" + httpPort,
			Handler:   r,
			TLSConfig: tlsConfig,
		}

		return server.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile)
	} else {
		return http.ListenAndServe(":"+httpPort, r)
	}
}

// startWebSocketServer initializes the WebSocket server
func (s *EnhancedBlockchainServer) startWebSocketServer(port string) {
	http.HandleFunc("/ws", s.handleWebSocketConnection)

	log.Printf("WebSocket server listening on port %s\n", port)

	if s.enableTLS {
		if err := http.ListenAndServeTLS(":"+port, s.tlsCertFile, s.tlsKeyFile, nil); err != nil {
			log.Printf("WebSocket server error: %v\n", err)
		}
	} else {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("WebSocket server error: %v\n", err)
		}
	}
}

// handleWebSocketConnection manages WebSocket client connections
func (s *EnhancedBlockchainServer) handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	// Register new client
	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	// Send initial stats
	s.sendStats(conn)

	// Handle client disconnection
	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		conn.Close()
	}()

	// Listen for messages from client (not used currently but could be for commands)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// handleBroadcasts sends messages to all connected WebSocket clients
func (s *EnhancedBlockchainServer) handleBroadcasts() {
	for message := range s.broadcast {
		s.clientsMutex.Lock()
		for client := range s.clients {
			err := client.WriteJSON(message)
			if err != nil {
				client.Close()
				delete(s.clients, client)
			}
		}
		s.clientsMutex.Unlock()
	}
}

// sendStats sends current blockchain stats to a specific client
func (s *EnhancedBlockchainServer) sendStats(conn *websocket.Conn) {
	stats := map[string]interface{}{
		"type":             "stats",
		"blockCount":       len(s.chain.Blocks),
		"transactionCount": s.txPool.Count(),
		"peerCount":        0, // To be implemented with P2P
		"nodeHealthy":      true,
	}

	conn.WriteJSON(stats)
}

// broadcastNewBlock notifies all clients about a new block
func (s *EnhancedBlockchainServer) broadcastNewBlock(block blockchain.Block) {
	s.broadcast <- map[string]interface{}{
		"type":  "new_block",
		"block": block,
	}
}

// broadcastNewTransaction notifies all clients about a new transaction
func (s *EnhancedBlockchainServer) broadcastNewTransaction(tx *blockchain.Transaction) {
	s.broadcast <- map[string]interface{}{
		"type":        "new_transaction",
		"transaction": tx,
	}
}

// broadcastContractDeployed notifies all clients about a new contract
func (s *EnhancedBlockchainServer) broadcastContractDeployed(contract interface{}) {
	s.broadcast <- map[string]interface{}{
		"type":     "contract_deployed",
		"contract": contract,
	}
}

// handleGetBlockchain returns the entire blockchain
func (s *EnhancedBlockchainServer) handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"blocks":     s.chain.Blocks,
		"difficulty": s.difficulty,
	}

	jsonResponse(w, response)
}

// handleGetBlocks returns all blocks or a subset with pagination
func (s *EnhancedBlockchainServer) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	// Could implement pagination here
	jsonResponse(w, map[string]interface{}{"blocks": s.chain.Blocks})
}

// handleGetBlock returns a specific block by hash
func (s *EnhancedBlockchainServer) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	for _, block := range s.chain.Blocks {
		if block.Hash == hash {
			jsonResponse(w, block)
			return
		}
	}

	http.Error(w, "Block not found", http.StatusNotFound)
}

// handleCreateTransaction adds a new transaction to the pool
func (s *EnhancedBlockchainServer) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var txData struct {
		From  string  `json:"from"`
		To    string  `json:"to"`
		Value float64 `json:"value"`
		Data  string  `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&txData); err != nil {
		http.Error(w, "Invalid transaction data", http.StatusBadRequest)
		return
	}

	// Create a new transaction
	tx := &blockchain.Transaction{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()), // Simple ID generation
		From:      txData.From,
		To:        txData.To,
		Data:      txData.Data,
		Value:     txData.Value,
		Timestamp: time.Now(),
		// Signature would be added in a real implementation
	}

	// Add to transaction pool
	if err := s.txPool.AddTransaction(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Record metrics
	s.metrics.TransactionProcessed(time.Millisecond * 10) // Placeholder processing time

	// Broadcast to WebSocket clients
	s.broadcastNewTransaction(tx)

	jsonResponse(w, map[string]string{"id": tx.ID, "status": "pending"})
}

// handleGetTransactions returns all transactions
func (s *EnhancedBlockchainServer) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would search transactions in blocks
	jsonResponse(w, map[string]interface{}{"transactions": s.txPool.GetAllTransactions()})
}

// handleGetTransaction returns a specific transaction by ID
func (s *EnhancedBlockchainServer) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	tx, err := s.txPool.GetTransaction(id)
	if err != nil {
		// Could search in confirmed transactions here
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, tx)
}

// handleGetPendingTransactions returns all pending transactions
func (s *EnhancedBlockchainServer) handleGetPendingTransactions(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]interface{}{"transactions": s.txPool.GetAllTransactions()})
}

// handleDeployContract deploys a new smart contract
func (s *EnhancedBlockchainServer) handleDeployContract(w http.ResponseWriter, r *http.Request) {
	var contractData struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&contractData); err != nil {
		http.Error(w, "Invalid contract data", http.StatusBadRequest)
		return
	}

	contractID := fmt.Sprintf("contract-%d", time.Now().UnixNano())
	var deployErr error
	var contractInfo interface{}

	switch contractData.Type {
	case "wasm":
		// In a real implementation, would save the WASM code to a file first
		// For now, just return an error as we can't deploy from code string directly
		http.Error(w, "WASM deployment from code string not supported", http.StatusNotImplemented)
		return

	case "lua":
		deployErr = s.luaEngine.DeployContract(contractID, contractData.Name, contractData.Code)
		if deployErr == nil {
			contract, _ := s.luaEngine.GetContract(contractID)
			contractInfo = map[string]interface{}{
				"id":   contractID,
				"name": contract.Name,
				"type": "lua",
			}
		}

	default:
		http.Error(w, "Unsupported contract type", http.StatusBadRequest)
		return
	}

	if deployErr != nil {
		http.Error(w, deployErr.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast to WebSocket clients
	s.broadcastContractDeployed(contractInfo)

	jsonResponse(w, map[string]interface{}{"id": contractID, "status": "deployed"})
}

// handleGetContracts returns all deployed contracts
func (s *EnhancedBlockchainServer) handleGetContracts(w http.ResponseWriter, r *http.Request) {
	wasmContracts := s.wasmEngine.ListContracts()
	luaContracts := s.luaEngine.ListContracts()

	// Convert to a common format
	contracts := make([]map[string]interface{}, 0)

	for _, c := range wasmContracts {
		contracts = append(contracts, map[string]interface{}{
			"id":   c.ID,
			"name": c.Name,
			"type": "wasm",
		})
	}

	for _, c := range luaContracts {
		contracts = append(contracts, map[string]interface{}{
			"id":   c.ID,
			"name": c.Name,
			"type": "lua",
		})
	}

	jsonResponse(w, map[string]interface{}{"contracts": contracts})
}

// handleGetContract returns a specific contract
func (s *EnhancedBlockchainServer) handleGetContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Try to find in WASM contracts
	wasmContract, err1 := s.wasmEngine.GetContract(id)
	if err1 == nil {
		jsonResponse(w, map[string]interface{}{
			"id":   wasmContract.ID,
			"name": wasmContract.Name,
			"type": "wasm",
		})
		return
	}

	// Try to find in Lua contracts
	luaContract, err2 := s.luaEngine.GetContract(id)
	if err2 == nil {
		jsonResponse(w, map[string]interface{}{
			"id":   luaContract.ID,
			"name": luaContract.Name,
			"type": "lua",
		})
		return
	}

	http.Error(w, "Contract not found", http.StatusNotFound)
}

// handleExecuteContract executes a function in a smart contract
func (s *EnhancedBlockchainServer) handleExecuteContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var execData struct {
		Function string        `json:"function"`
		Params   []interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&execData); err != nil {
		http.Error(w, "Invalid execution data", http.StatusBadRequest)
		return
	}

	// Try to execute WASM contract
	_, err1 := s.wasmEngine.GetContract(id)
	if err1 == nil {
		result, err := s.wasmEngine.ExecuteContract(id, execData.Function, execData.Params...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, map[string]interface{}{"result": result})
		return
	}

	// Try to execute Lua contract
	_, err2 := s.luaEngine.GetContract(id)
	if err2 == nil {
		result, err := s.luaEngine.ExecuteContract(id, execData.Function, execData.Params...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, map[string]interface{}{"result": result})
		return
	}

	http.Error(w, "Contract not found", http.StatusNotFound)
}

// jsonResponse sends a JSON response with the given data
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
