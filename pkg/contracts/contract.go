package contracts

// ContractEngine defines the interface for smart contract execution engines
type ContractEngine interface {
	// DeployContract deploys a new contract
	// For WASM engine, code is a file path
	// For Lua engine, code is the actual Lua code
	DeployContract(id string, name string, code string) error

	// ExecuteContract runs a function in a contract with the given parameters
	ExecuteContract(contractID string, functionName string, params ...interface{}) (interface{}, error)

	// GetContract retrieves contract information by ID
	GetContract(id string) (interface{}, error)

	// RemoveContract deletes a contract
	RemoveContract(id string) error
}

// ContractInfo contains common contract metadata
type ContractInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "wasm" or "lua"
}
