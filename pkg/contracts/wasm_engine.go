package contracts

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/wasmerio/wasmer-go/wasmer"
)

// WASMEngine provides WebAssembly-based smart contract execution
type WASMEngine struct {
	contracts map[string]*Contract
	mutex     sync.RWMutex
}

// Contract represents a compiled WASM smart contract
type Contract struct {
	ID        string
	Name      string
	Code      []byte
	Instance  *wasmer.Instance
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewWASMEngine creates a new WebAssembly smart contract engine
func NewWASMEngine() *WASMEngine {
	return &WASMEngine{
		contracts: make(map[string]*Contract),
	}
}

// DeployContract loads and compiles a WASM contract from a file
func (e *WASMEngine) DeployContract(id, name, filePath string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Read the WASM file
	wasmBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file: %w", err)
	}

	// Create a new WebAssembly Instance
	store := wasmer.NewStore(wasmer.NewEngine())
	module, err := wasmer.NewModule(store, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Instantiate the WebAssembly module
	instance, err := wasmer.NewInstance(module, wasmer.NewImportObject())
	if err != nil {
		return fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	// Store the contract
	e.contracts[id] = &Contract{
		ID:        id,
		Name:      name,
		Code:      wasmBytes,
		Instance:  instance,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

// ExecuteContract runs a function in the specified contract
func (e *WASMEngine) ExecuteContract(contractID, functionName string, params ...interface{}) (interface{}, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// Get the contract
	contract, exists := e.contracts[contractID]
	if !exists {
		return nil, errors.New("contract not found")
	}

	// Get the function from the instance
	func_, err := contract.Instance.Exports.GetFunction(functionName)
	if err != nil {
		return nil, fmt.Errorf("function not found: %w", err)
	}

	// Execute the function
	result, err := func_(params...)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return result, nil
}

// GetContract returns a contract by ID
func (e *WASMEngine) GetContract(id string) (*Contract, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	contract, exists := e.contracts[id]
	if !exists {
		return nil, errors.New("contract not found")
	}

	return contract, nil
}

// ListContracts returns all deployed contracts
func (e *WASMEngine) ListContracts() []*Contract {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	contracts := make([]*Contract, 0, len(e.contracts))
	for _, contract := range e.contracts {
		contracts = append(contracts, contract)
	}

	return contracts
}

// RemoveContract deletes a contract by ID
func (e *WASMEngine) RemoveContract(id string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	contract, exists := e.contracts[id]
	if !exists {
		return errors.New("contract not found")
	}

	// Close the WebAssembly instance
	contract.Instance.Close()

	// Remove the contract from the map
	delete(e.contracts, id)

	return nil
}
