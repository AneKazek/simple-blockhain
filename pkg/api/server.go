package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/anekazek/simple-blockchain/pkg/blockchain"
)

// BlockchainServer handles HTTP requests for blockchain operations
type BlockchainServer struct {
	chain      *blockchain.Chain
	difficulty int
}

// NewBlockchainServer creates a new server with the given blockchain
func NewBlockchainServer(chain *blockchain.Chain, difficulty int) *BlockchainServer {
	return &BlockchainServer{
		chain:      chain,
		difficulty: difficulty,
	}
}

// Start initializes the HTTP server and routes
func (s *BlockchainServer) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleGetBlockchain)
	mux.HandleFunc("/write", s.handleWriteBlock)

	log.Printf("Server listening on port %s\n", port)
	return http.ListenAndServe(":"+port, mux)
}

// handleGetBlockchain returns the entire blockchain
func (s *BlockchainServer) handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(s.chain.GetBlocks(), "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// handleWriteBlock adds a new block to the blockchain
func (s *BlockchainServer) handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Data string `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := s.chain.AddBlock(data.Data, s.difficulty)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, struct{ Error string }{Error: err.Error()})
		return
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

// respondWithJSON is a helper function to send JSON responses
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
