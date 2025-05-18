package consensus

import (
	"strings"

	"github.com/anekazek/simple-blockchain/pkg/blockchain"
)

// ProofOfWork implements the Proof of Work consensus algorithm
type ProofOfWork struct {
	Difficulty int
}

// NewProofOfWork creates a new PoW consensus with the specified difficulty
func NewProofOfWork(difficulty int) *ProofOfWork {
	return &ProofOfWork{
		Difficulty: difficulty,
	}
}

// ValidateBlock checks if a block's hash meets the difficulty requirement
func (pow *ProofOfWork) ValidateBlock(block blockchain.Block) bool {
	prefix := strings.Repeat("0", pow.Difficulty)
	return strings.HasPrefix(block.Hash, prefix)
}

// SetDifficulty changes the mining difficulty
func (pow *ProofOfWork) SetDifficulty(difficulty int) {
	pow.Difficulty = difficulty
}

// GetDifficulty returns the current mining difficulty
func (pow *ProofOfWork) GetDifficulty() int {
	return pow.Difficulty
}
