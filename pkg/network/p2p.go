package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/anekazek/simple-blockchain/pkg/blockchain"
)

// Peer represents a node in the P2P network
type Peer struct {
	Address  string
	LastSeen time.Time
}

// P2PServer manages peer-to-peer communication between blockchain nodes
type P2PServer struct {
	chain       *blockchain.Chain
	peers       map[string]Peer
	peersMutex  *sync.Mutex
	port        string
	knownBlocks map[string]bool // Track blocks we've already seen by hash
}

// NewP2PServer creates a new P2P server for the given blockchain
func NewP2PServer(chain *blockchain.Chain, port string) *P2PServer {
	return &P2PServer{
		chain:       chain,
		peers:       make(map[string]Peer),
		peersMutex:  &sync.Mutex{},
		port:        port,
		knownBlocks: make(map[string]bool),
	}
}

// RegisterRoutes adds P2P endpoints to the HTTP server
func (p *P2PServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/peers", p.handlePeers)
	mux.HandleFunc("/register-peer", p.handleRegisterPeer)
	mux.HandleFunc("/sync", p.handleSync)
	mux.HandleFunc("/broadcast-block", p.handleBroadcastBlock)
}

// Start begins the P2P server operations
func (p *P2PServer) Start() {
	// Start periodic peer discovery and chain synchronization
	go p.discoverPeers()
	go p.syncBlockchain()
}

// AddPeer adds a new peer to the network
func (p *P2PServer) AddPeer(address string) {
	p.peersMutex.Lock()
	defer p.peersMutex.Unlock()

	p.peers[address] = Peer{
		Address:  address,
		LastSeen: time.Now(),
	}
	log.Printf("Added peer: %s\n", address)
}

// BroadcastBlock sends a new block to all peers
func (p *P2PServer) BroadcastBlock(block blockchain.Block) {
	p.peersMutex.Lock()
	peers := make([]string, 0, len(p.peers))
	for addr := range p.peers {
		peers = append(peers, addr)
	}
	p.peersMutex.Unlock()

	for _, peer := range peers {
		go func(address string) {
			url := fmt.Sprintf("http://%s/broadcast-block", address)
			blockData, _ := json.Marshal(block)
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(blockData))
			if err != nil {
				log.Printf("Failed to broadcast block to %s: %v\n", address, err)
				return
			}
			defer resp.Body.Close()
		}(peer)
	}
}

// discoverPeers periodically looks for new peers
func (p *P2PServer) discoverPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		p.peersMutex.Lock()
		peers := make([]string, 0, len(p.peers))
		for addr := range p.peers {
			peers = append(peers, addr)
		}
		p.peersMutex.Unlock()

		// Ask each peer for their peers
		for _, peer := range peers {
			go func(address string) {
				url := fmt.Sprintf("http://%s/peers", address)
				resp, err := http.Get(url)
				if err != nil {
					log.Printf("Failed to get peers from %s: %v\n", address, err)
					return
				}
				defer resp.Body.Close()

				var peerList []string
				if err := json.NewDecoder(resp.Body).Decode(&peerList); err != nil {
					log.Printf("Failed to decode peers from %s: %v\n", address, err)
					return
				}

				// Register new peers
				for _, newPeer := range peerList {
					if newPeer != p.port && newPeer != address {
						p.AddPeer(newPeer)
						// Register ourselves with the new peer
						p.registerWithPeer(newPeer)
					}
				}
			}(peer)
		}
	}
}

// syncBlockchain periodically syncs the blockchain with peers
func (p *P2PServer) syncBlockchain() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		p.peersMutex.Lock()
		peers := make([]string, 0, len(p.peers))
		for addr := range p.peers {
			peers = append(peers, addr)
		}
		p.peersMutex.Unlock()

		// Sync with each peer
		for _, peer := range peers {
			go func(address string) {
				url := fmt.Sprintf("http://%s/", address)
				resp, err := http.Get(url)
				if err != nil {
					log.Printf("Failed to sync with %s: %v\n", address, err)
					return
				}
				defer resp.Body.Close()

				var blocks []blockchain.Block
				if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
					log.Printf("Failed to decode blockchain from %s: %v\n", address, err)
					return
				}

				// Replace our chain if the peer has a longer valid chain
				if len(blocks) > len(p.chain.GetBlocks()) {
					if p.chain.ReplaceChain(blocks) {
						log.Printf("Blockchain replaced with longer chain from %s\n", address)
					}
				}
			}(peer)
		}
	}
}

// registerWithPeer registers this node with another peer
func (p *P2PServer) registerWithPeer(peerAddr string) {
	url := fmt.Sprintf("http://%s/register-peer", peerAddr)
	data := map[string]string{"address": p.port}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to register with peer %s: %v\n", peerAddr, err)
		return
	}
	defer resp.Body.Close()
}

// HTTP Handlers

func (p *P2PServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	p.peersMutex.Lock()
	peerAddresses := make([]string, 0, len(p.peers))
	for addr := range p.peers {
		peerAddresses = append(peerAddresses, addr)
	}
	p.peersMutex.Unlock()

	json.NewEncoder(w).Encode(peerAddresses)
}

func (p *P2PServer) handleRegisterPeer(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	address, ok := data["address"]
	if !ok {
		http.Error(w, "Missing peer address", http.StatusBadRequest)
		return
	}

	p.AddPeer(address)
	w.WriteHeader(http.StatusOK)
}

func (p *P2PServer) handleSync(w http.ResponseWriter, r *http.Request) {
	blocks := p.chain.GetBlocks()
	json.NewEncoder(w).Encode(blocks)
}

func (p *P2PServer) handleBroadcastBlock(w http.ResponseWriter, r *http.Request) {
	var block blockchain.Block
	if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if we've already seen this block
	if p.knownBlocks[block.Hash] {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mark this block as seen
	p.knownBlocks[block.Hash] = true

	// Validate and add the block to our chain if valid
	if blockchain.IsBlockValid(block, p.chain.GetLatestBlock()) {
		p.chain.ReplaceChain(append(p.chain.GetBlocks(), block))
		log.Printf("Added new block from peer: %s\n", block.Hash)

		// Forward the block to other peers (except the one who sent it)
		io.Copy(io.Discard, r.Body) // Drain the body
		peerAddr := r.Header.Get("X-Forwarded-For")
		if peerAddr == "" {
			peerAddr = r.RemoteAddr
		}

		p.peersMutex.Lock()
		peers := make([]string, 0, len(p.peers))
		for addr := range p.peers {
			if addr != peerAddr {
				peers = append(peers, addr)
			}
		}
		p.peersMutex.Unlock()

		for _, peer := range peers {
			go func(address string) {
				url := fmt.Sprintf("http://%s/broadcast-block", address)
				blockData, _ := json.Marshal(block)
				resp, err := http.Post(url, "application/json", bytes.NewBuffer(blockData))
				if err != nil {
					log.Printf("Failed to forward block to %s: %v\n", address, err)
					return
				}
				defer resp.Body.Close()
			}(peer)
		}
	}

	w.WriteHeader(http.StatusOK)
}
