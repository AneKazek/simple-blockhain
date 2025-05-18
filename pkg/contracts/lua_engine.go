package contracts

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// LuaEngine provides Lua-based smart contract execution
type LuaEngine struct {
	contracts map[string]*LuaContract
	mutex     sync.RWMutex
}

// LuaContract represents a Lua smart contract
type LuaContract struct {
	ID        string
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewLuaEngine creates a new Lua smart contract engine
func NewLuaEngine() *LuaEngine {
	return &LuaEngine{
		contracts: make(map[string]*LuaContract),
	}
}

// DeployContract loads and registers a Lua contract
func (e *LuaEngine) DeployContract(id, name, code string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Validate the Lua code by attempting to load it
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(code)
	if err != nil {
		return fmt.Errorf("invalid Lua code: %w", err)
	}

	// Store the contract
	e.contracts[id] = &LuaContract{
		ID:        id,
		Name:      name,
		Code:      code,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

// ExecuteContract runs a function in the specified Lua contract
func (e *LuaEngine) ExecuteContract(contractID, functionName string, params ...interface{}) (interface{}, error) {
	e.mutex.RLock()
	contract, exists := e.contracts[contractID]
	if !exists {
		e.mutex.RUnlock()
		return nil, errors.New("contract not found")
	}
	code := contract.Code
	e.mutex.RUnlock()

	// Create a new Lua state for execution
	L := lua.NewState()
	defer L.Close()

	// Load the contract code
	err := L.DoString(code)
	if err != nil {
		return nil, fmt.Errorf("failed to load contract: %w", err)
	}

	// Get the function
	luaFunc := L.GetGlobal(functionName)
	if luaFunc.Type() != lua.LTFunction {
		return nil, fmt.Errorf("function '%s' not found in contract", functionName)
	}

	// Convert Go params to Lua values
	luaParams := make([]lua.LValue, len(params))
	for i, param := range params {
		switch v := param.(type) {
		case string:
			luaParams[i] = lua.LString(v)
		case int:
			luaParams[i] = lua.LNumber(v)
		case float64:
			luaParams[i] = lua.LNumber(v)
		case bool:
			luaParams[i] = lua.LBool(v)
		default:
			return nil, fmt.Errorf("unsupported parameter type: %T", param)
		}
	}

	// Call the function
	err = L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1,
		Protect: true,
	}, luaParams...)

	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	// Get the result
	result := L.Get(-1)
	L.Pop(1)

	// Convert Lua value to Go value
	switch result.Type() {
	case lua.LTNil:
		return nil, nil
	case lua.LTBool:
		return lua.LVAsBool(result), nil
	case lua.LTNumber:
		return float64(result.(lua.LNumber)), nil
	case lua.LTString:
		return string(result.(lua.LString)), nil
	default:
		return nil, fmt.Errorf("unsupported return type: %s", result.Type().String())
	}
}

// GetContract returns a contract by ID
func (e *LuaEngine) GetContract(id string) (*LuaContract, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	contract, exists := e.contracts[id]
	if !exists {
		return nil, errors.New("contract not found")
	}

	return contract, nil
}

// ListContracts returns all deployed contracts
func (e *LuaEngine) ListContracts() []*LuaContract {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	contracts := make([]*LuaContract, 0, len(e.contracts))
	for _, contract := range e.contracts {
		contracts = append(contracts, contract)
	}

	return contracts
}

// RemoveContract deletes a contract by ID
func (e *LuaEngine) RemoveContract(id string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	_, exists := e.contracts[id]
	if !exists {
		return errors.New("contract not found")
	}

	// Remove the contract from the map
	delete(e.contracts, id)

	return nil
}
