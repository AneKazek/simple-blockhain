package consensus

import (
	"github.com/anekazek/simple-blockchain/pkg/blockchain"
)

// Algorithm defines the interface for consensus algorithms
type Algorithm interface {
	// ValidateBlock checks if a block meets the consensus requirements
	ValidateBlock(block blockchain.Block) bool

	// SetDifficulty changes the consensus difficulty parameter
	SetDifficulty(difficulty int)

	// GetDifficulty returns the current difficulty parameter
	GetDifficulty() int
}
