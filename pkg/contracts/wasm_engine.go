package contracts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// WASMEngine provides WebAssembly-based smart contract execution
type WASMEngine struct {
	contracts map[string]*Contract
	runtime   wazero.Runtime
	mutex     sync.RWMutex
	ctx       context.Context
}

// Contract represents a compiled WASM smart contract
type Contract struct {
	ID        string
	Name      string
	Code      []byte
	Module    api.Module
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewWASMEngine creates a new WebAssembly smart contract engine
func NewWASMEngine() *WASMEngine {
	ctx := context.Background()
	// Create a new WebAssembly Runtime
	runtime := wazero.NewRuntime(ctx)

	return &WASMEngine{
		contracts: make(map[string]*Contract),
		runtime:   runtime,
		ctx:       ctx,
	}
}

// DeployContract loads and compiles a WASM contract from a file
func (e *WASMEngine) DeployContract(id, name, filePath string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Read the WASM file
	wasmBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file: %w", err)
	}

	// Compile the WebAssembly module
	module, err := e.runtime.CompileModule(e.ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Instantiate the WebAssembly module
	instance, err := e.runtime.InstantiateModule(e.ctx, module, wazero.NewModuleConfig())
	if err != nil {
		return fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	// Store the contract
	e.contracts[id] = &Contract{
		ID:        id,
		Name:      name,
		Code:      wasmBytes,
		Module:    instance,
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

	// Get the function from the module
	fn := contract.Module.ExportedFunction(functionName)
	if fn == nil {
		return nil, fmt.Errorf("function not found: %s", functionName)
	}

	// Convert params to wazero format
	wasmParams := make([]uint64, 0, len(params))
	for _, param := range params {
		switch v := param.(type) {
		case int:
			wasmParams = append(wasmParams, uint64(v))
		case int32:
			wasmParams = append(wasmParams, uint64(v))
		case int64:
			wasmParams = append(wasmParams, uint64(v))
		case uint:
			wasmParams = append(wasmParams, uint64(v))
		case uint32:
			wasmParams = append(wasmParams, uint64(v))
		case uint64:
			wasmParams = append(wasmParams, v)
		case float32:
			wasmParams = append(wasmParams, uint64(v))
		case float64:
			wasmParams = append(wasmParams, uint64(v))
		default:
			return nil, fmt.Errorf("unsupported parameter type: %T", param)
		}
	}

	// Execute the function
	results, err := fn.Call(e.ctx, wasmParams...)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0], nil
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

	// Close the WebAssembly module
	err := contract.Module.Close(e.ctx)
	if err != nil {
		return fmt.Errorf("failed to close module: %w", err)
	}

	// Remove the contract from the map
	delete(e.contracts, id)

	return nil
}
