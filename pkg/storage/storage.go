package storage

import (
	"github.com/anekazek/simple-blockchain/pkg/blockchain"
)

// BlockchainStore defines the interface for blockchain storage implementations
type BlockchainStore interface {
	// Initialize prepares the storage for use
	Initialize() error

	// SaveBlock persists a block to storage
	SaveBlock(block blockchain.Block) error

	// GetBlock retrieves a block by its hash
	GetBlock(hash string) (blockchain.Block, error)

	// GetBlockByIndex retrieves a block by its index
	GetBlockByIndex(index int) (blockchain.Block, error)

	// GetAllBlocks retrieves all blocks from storage
	GetAllBlocks() ([]blockchain.Block, error)

	// GetLatestBlock retrieves the most recent block
	GetLatestBlock() (blockchain.Block, error)

	// Close closes the storage connection
	Close() error
}
