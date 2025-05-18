package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/anekazek/simple-blockchain/pkg/blockchain"
	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBStore implements BlockchainStore using LevelDB
type LevelDBStore struct {
	db        *leveldb.DB
	dbPath    string
	lastIndex int
}

// NewLevelDBStore creates a new LevelDB-backed blockchain store
func NewLevelDBStore(dbPath string) *LevelDBStore {
	return &LevelDBStore{
		dbPath:    dbPath,
		lastIndex: -1,
	}
}

// Initialize opens the database connection
func (s *LevelDBStore) Initialize() error {
	db, err := leveldb.OpenFile(s.dbPath, nil)
	if err != nil {
		return fmt.Errorf("failed to open leveldb: %w", err)
	}
	s.db = db

	// Find the last index
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		if len(key) > 5 && key[:5] == "index" {
			indexStr := key[5:]
			index, err := strconv.Atoi(indexStr)
			if err == nil && index > s.lastIndex {
				s.lastIndex = index
			}
		}
	}

	return nil
}

// SaveBlock persists a block to the database
func (s *LevelDBStore) SaveBlock(block blockchain.Block) error {
	if s.db == nil {
		return errors.New("database not initialized")
	}

	// Store by hash
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// Store by hash
	err = s.db.Put([]byte("hash"+block.Hash), blockData, nil)
	if err != nil {
		return fmt.Errorf("failed to store block by hash: %w", err)
	}

	// Store by index
	err = s.db.Put([]byte("index"+strconv.Itoa(block.Index)), blockData, nil)
	if err != nil {
		return fmt.Errorf("failed to store block by index: %w", err)
	}

	// Update last index if needed
	if block.Index > s.lastIndex {
		s.lastIndex = block.Index
		// Store the latest block hash
		err = s.db.Put([]byte("latest"), []byte(block.Hash), nil)
		if err != nil {
			return fmt.Errorf("failed to update latest block: %w", err)
		}
	}

	return nil
}

// GetBlock retrieves a block by its hash
func (s *LevelDBStore) GetBlock(hash string) (blockchain.Block, error) {
	if s.db == nil {
		return blockchain.Block{}, errors.New("database not initialized")
	}

	data, err := s.db.Get([]byte("hash"+hash), nil)
	if err != nil {
		return blockchain.Block{}, fmt.Errorf("block not found: %w", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return blockchain.Block{}, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return block, nil
}

// GetBlockByIndex retrieves a block by its index
func (s *LevelDBStore) GetBlockByIndex(index int) (blockchain.Block, error) {
	if s.db == nil {
		return blockchain.Block{}, errors.New("database not initialized")
	}

	data, err := s.db.Get([]byte("index"+strconv.Itoa(index)), nil)
	if err != nil {
		return blockchain.Block{}, fmt.Errorf("block not found: %w", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return blockchain.Block{}, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return block, nil
}

// GetAllBlocks retrieves all blocks from storage
func (s *LevelDBStore) GetAllBlocks() ([]blockchain.Block, error) {
	if s.db == nil {
		return nil, errors.New("database not initialized")
	}

	blocks := make([]blockchain.Block, s.lastIndex+1)
	for i := 0; i <= s.lastIndex; i++ {
		block, err := s.GetBlockByIndex(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at index %d: %w", i, err)
		}
		blocks[i] = block
	}

	return blocks, nil
}

// GetLatestBlock retrieves the most recent block
func (s *LevelDBStore) GetLatestBlock() (blockchain.Block, error) {
	if s.db == nil {
		return blockchain.Block{}, errors.New("database not initialized")
	}

	// Get the hash of the latest block
	hashBytes, err := s.db.Get([]byte("latest"), nil)
	if err != nil {
		return blockchain.Block{}, fmt.Errorf("latest block not found: %w", err)
	}

	// Get the block by hash
	return s.GetBlock(string(hashBytes))
}

// Close closes the database connection
func (s *LevelDBStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
