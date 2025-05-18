package consensus

import (
	"math/rand"
	"time"

	"github.com/anekazek/simple-blockchain/pkg/blockchain"
)

// ProofOfStake implements a basic Proof of Stake consensus algorithm
type ProofOfStake struct {
	Difficulty int
	Stakers    map[string]int
	rand       *rand.Rand
}

// NewProofOfStake creates a new PoS consensus with the specified difficulty
func NewProofOfStake(difficulty int) *ProofOfStake {
	return &ProofOfStake{
		Difficulty: difficulty,
		Stakers:    make(map[string]int),
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddStaker adds a new staker with the specified stake amount
func (pos *ProofOfStake) AddStaker(address string, stake int) {
	pos.Stakers[address] = stake
}

// SelectValidator chooses a validator based on their stake
func (pos *ProofOfStake) SelectValidator() string {
	totalStake := 0
	for _, stake := range pos.Stakers {
		totalStake += stake
	}

	if totalStake == 0 {
		return ""
	}

	// Select a random point in the stake space
	selection := pos.rand.Intn(totalStake)

	// Find which staker owns that point
	currentPosition := 0
	for address, stake := range pos.Stakers {
		currentPosition += stake
		if selection < currentPosition {
			return address
		}
	}

	return ""
}

// ValidateBlock checks if a block is valid according to PoS rules
// In a real implementation, this would verify the validator's signature
func (pos *ProofOfStake) ValidateBlock(block blockchain.Block) bool {
	// In a real implementation, we would verify:
	// 1. The block is signed by a valid validator
	// 2. The validator was selected for this time slot
	// 3. The validator has sufficient stake

	// For this simple implementation, we'll just return true
	// as if the block was properly validated
	return true
}

// SetDifficulty changes the consensus parameter (not directly used in PoS)
func (pos *ProofOfStake) SetDifficulty(difficulty int) {
	pos.Difficulty = difficulty
}

// GetDifficulty returns the current difficulty parameter
func (pos *ProofOfStake) GetDifficulty() int {
	return pos.Difficulty
}
