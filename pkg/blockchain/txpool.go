package blockchain

import (
	"errors"
	"sync"
	"time"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Data      string    `json:"data"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"`
}

// TransactionPool manages pending transactions
type TransactionPool struct {
	pendingTransactions map[string]*Transaction
	mutex               sync.RWMutex
	maxPoolSize         int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool(maxPoolSize int) *TransactionPool {
	if maxPoolSize <= 0 {
		maxPoolSize = 1000 // Default max pool size
	}

	return &TransactionPool{
		pendingTransactions: make(map[string]*Transaction),
		maxPoolSize:         maxPoolSize,
	}
}

// AddTransaction adds a transaction to the pool
func (tp *TransactionPool) AddTransaction(tx *Transaction) error {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	// Check if pool is full
	if len(tp.pendingTransactions) >= tp.maxPoolSize {
		return errors.New("transaction pool is full")
	}

	// Check if transaction already exists
	if _, exists := tp.pendingTransactions[tx.ID]; exists {
		return errors.New("transaction already exists in pool")
	}

	// Add transaction to pool
	tp.pendingTransactions[tx.ID] = tx
	return nil
}

// GetTransaction retrieves a transaction from the pool
func (tp *TransactionPool) GetTransaction(txID string) (*Transaction, error) {
	tp.mutex.RLock()
	defer tp.mutex.RUnlock()

	tx, exists := tp.pendingTransactions[txID]
	if !exists {
		return nil, errors.New("transaction not found in pool")
	}

	return tx, nil
}

// GetAllTransactions returns all transactions in the pool
func (tp *TransactionPool) GetAllTransactions() []*Transaction {
	tp.mutex.RLock()
	defer tp.mutex.RUnlock()

	transactions := make([]*Transaction, 0, len(tp.pendingTransactions))
	for _, tx := range tp.pendingTransactions {
		transactions = append(transactions, tx)
	}

	return transactions
}

// RemoveTransaction removes a transaction from the pool
func (tp *TransactionPool) RemoveTransaction(txID string) error {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	if _, exists := tp.pendingTransactions[txID]; !exists {
		return errors.New("transaction not found in pool")
	}

	delete(tp.pendingTransactions, txID)
	return nil
}

// GetBatch retrieves a batch of transactions for block creation
func (tp *TransactionPool) GetBatch(maxCount int) []*Transaction {
	tp.mutex.RLock()
	defer tp.mutex.RUnlock()

	count := 0
	transactions := make([]*Transaction, 0, maxCount)

	for _, tx := range tp.pendingTransactions {
		if count >= maxCount {
			break
		}
		transactions = append(transactions, tx)
		count++
	}

	return transactions
}

// RemoveBatch removes a batch of transactions from the pool
func (tp *TransactionPool) RemoveBatch(txIDs []string) {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	for _, id := range txIDs {
		delete(tp.pendingTransactions, id)
	}
}

// Clear empties the transaction pool
func (tp *TransactionPool) Clear() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	tp.pendingTransactions = make(map[string]*Transaction)
}

// Count returns the number of transactions in the pool
func (tp *TransactionPool) Count() int {
	tp.mutex.RLock()
	defer tp.mutex.RUnlock()

	return len(tp.pendingTransactions)
}
