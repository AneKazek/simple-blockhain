package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index      int    `json:"index"`
	Timestamp  string `json:"timestamp"`
	Data       string `json:"data"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prevHash"`
	Difficulty int    `json:"difficulty"`
	Nonce      string `json:"nonce"`
}

// CalculateHash is a simple SHA256 hashing function
func CalculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Data + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// GenerateBlock creates a new block using previous block's hash
func GenerateBlock(oldBlock Block, data string, difficulty int) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Data = data
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		newBlockHash := CalculateHash(newBlock)
		if !IsHashValid(newBlockHash, newBlock.Difficulty) {
			fmt.Printf("\r%s", newBlockHash)
			continue
		}
		fmt.Println()
		newBlock.Hash = newBlockHash
		break
	}

	return newBlock, nil
}

// IsBlockValid makes sure block is valid by checking index
// and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if CalculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// IsHashValid checks if hash meets difficulty requirement
func IsHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// CreateGenesisBlock creates the first block in the blockchain
func CreateGenesisBlock() Block {
	t := time.Now()
	genesisBlock := Block{
		Index:      0,
		Timestamp:  t.String(),
		Data:       "Genesis Block",
		Difficulty: 1,
		Nonce:      "",
		PrevHash:   "",
	}
	genesisBlock.Hash = CalculateHash(genesisBlock)
	return genesisBlock
}
