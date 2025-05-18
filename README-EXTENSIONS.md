# Simple Blockchain Extensions

This document describes the extensions implemented for the Simple Blockchain project, including smart contract support, transaction pool management, web dashboard, and enhanced API with WebSocket support.

## Smart Contract Engine

Two smart contract engines have been implemented:

### WebAssembly (WASM) Engine

The WASM engine allows deploying and executing WebAssembly smart contracts, providing a secure and efficient execution environment.

**Features:**
- Deploy WASM contracts from files
- Execute contract functions with parameters
- Manage contract lifecycle

**Dependencies:**
- [wazero](https://github.com/tetratelabs/wazero) - Zero dependency WebAssembly runtime for Go

### Lua Engine

The Lua engine provides a simpler scripting option for smart contracts using the Lua programming language.

**Features:**
- Deploy Lua contracts directly from code strings
- Execute Lua functions with automatic type conversion
- Lightweight and easy to use

**Dependencies:**
- [gopher-lua](https://github.com/yuin/gopher-lua) - Lua VM implementation in Go

## Transaction Pool Management

A transaction pool has been implemented to manage pending transactions before they're added to blocks.

**Features:**
- Add, retrieve, and remove transactions
- Get transaction batches for block creation
- Configurable pool size

## Web Dashboard

A responsive web dashboard has been created to monitor and interact with the blockchain.

**Features:**
- Real-time updates via WebSocket
- View blockchain statistics
- Browse blocks and transactions
- Create and submit transactions
- Deploy and execute smart contracts

## Enhanced API Server

The API server has been enhanced with WebSocket support and HTTPS/TLS authentication.

**Features:**
- RESTful API for blockchain interaction
- WebSocket server for real-time updates
- HTTPS/TLS support for secure communication
- Smart contract deployment and execution endpoints

## Configuration

The following environment variables can be used to configure the application:

- `BLOCKCHAIN_DIFFICULTY` - Mining difficulty (default: 1)
- `TX_POOL_SIZE` - Transaction pool capacity (default: 1000)
- `HTTP_PORT` - HTTP API port (default: 8080)
- `WS_PORT` - WebSocket server port (default: 8081)
- `METRICS_PORT` - Prometheus metrics port (default: 9090)
- `TLS_CERT_FILE` - Path to TLS certificate file (optional)
- `TLS_KEY_FILE` - Path to TLS key file (optional)

## Usage

### Starting the Application

```bash
# Basic start
go run main.go

# With custom configuration
BLOCKCHAIN_DIFFICULTY=2 TX_POOL_SIZE=2000 HTTP_PORT=8000 WS_PORT=8001 go run main.go
```

### Accessing the Dashboard

Open your browser and navigate to:

```
http://localhost:8080
```

### Accessing Metrics

Prometheus metrics are available at:

```
http://localhost:9090/metrics
```

### API Endpoints

#### Blockchain
- `GET /api/blockchain` - Get the entire blockchain
- `GET /api/blocks` - Get all blocks
- `GET /api/blocks/{hash}` - Get a specific block by hash

#### Transactions
- `POST /api/transactions` - Create a new transaction
- `GET /api/transactions` - Get all transactions
- `GET /api/transactions/{id}` - Get a specific transaction by ID
- `GET /api/transactions/pending` - Get all pending transactions

#### Smart Contracts
- `POST /api/contracts` - Deploy a new smart contract
- `GET /api/contracts` - Get all deployed contracts
- `GET /api/contracts/{id}` - Get a specific contract by ID
- `POST /api/contracts/{id}/execute` - Execute a function in a smart contract

## Dependencies

To install the required dependencies:

```bash
go get github.com/tetratelabs/wazero
go get github.com/yuin/gopher-lua
go get github.com/gorilla/mux
go get github.com/gorilla/websocket
go get github.com/prometheus/client_golang/prometheus
```

## Security Considerations

- For production use, always enable TLS by providing certificate and key files
- Implement proper authentication and authorization mechanisms
- Consider using a more sophisticated transaction signature scheme
- Implement rate limiting for API endpoints