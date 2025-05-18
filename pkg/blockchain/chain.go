package blockchain

import (
	"sync"
)

// Chain represents the blockchain and provides methods to interact with it
type Chain struct {
	Blocks []Block
	mutex  *sync.Mutex
}

// NewBlockchain creates a new blockchain with a genesis block
func NewBlockchain() *Chain {
	genesisBlock := CreateGenesisBlock()
	return &Chain{
		Blocks: []Block{genesisBlock},
		mutex:  &sync.Mutex{},
	}
}

// AddBlock adds a new block to the blockchain if it's valid
func (bc *Chain) AddBlock(data string, difficulty int) (Block, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	newBlock, err := GenerateBlock(bc.Blocks[len(bc.Blocks)-1], data, difficulty)
	if err != nil {
		return Block{}, err
	}

	if IsBlockValid(newBlock, bc.Blocks[len(bc.Blocks)-1]) {
		bc.Blocks = append(bc.Blocks, newBlock)
	}

	return newBlock, nil
}

// GetLatestBlock returns the most recent block in the chain
func (bc *Chain) GetLatestBlock() Block {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	return bc.Blocks[len(bc.Blocks)-1]
}

// ReplaceChain replaces our chain with a new one if it's longer and valid
func (bc *Chain) ReplaceChain(newChain []Block) bool {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(newChain) <= len(bc.Blocks) {
		return false
	}

	// Validate the new chain
	for i := 1; i < len(newChain); i++ {
		if !IsBlockValid(newChain[i], newChain[i-1]) {
			return false
		}
	}

	bc.Blocks = newChain
	return true
}

// GetBlocks returns all blocks in the chain
func (bc *Chain) GetBlocks() []Block {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	return bc.Blocks
}
