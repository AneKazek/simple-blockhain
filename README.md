# Simple Blockchain

A lightweight blockchain implementation in Go, based on the "Code your own blockchain in less than 200 lines of Go" tutorial, with extended enterprise-grade features.

## Core Features

- **Basic Blockchain**: Block structure, hashing, mining, and HTTP server in under 200 lines
- **Proof-of-Work Consensus**: Simple mining with configurable difficulty
- **RESTful API**: HTTP endpoints for blockchain queries and adding new blocks

## Planned Extensions

- **P2P Networking**: Multi-node synchronization with auto peer discovery
- **Multiple Consensus Algorithms**: Configurable PoW and PoS
- **Smart Contract Engine**: WASM or embedded Lua support
- **Enhanced API**: Full RESTful API for ledger queries and transaction management
- **Persistent Storage**: LevelDB or BoltDB integration
- **Web Dashboard**: Interactive UI with WebSocket/GraphQL updates
- **Security**: HTTPS/TLS mutual-authentication for P2P and API
- **Monitoring**: Prometheus metrics for TPS, latency, and node health

## Project Structure

```
.
├── main.go                 # Core blockchain implementation (<200 lines)
├── cmd/                    # Command-line applications
├── pkg/                    # Reusable packages
│   ├── blockchain/         # Core blockchain logic
│   ├── consensus/          # Consensus algorithms (PoW, PoS)
│   ├── network/            # P2P networking
│   ├── api/                # RESTful API
│   ├── storage/            # Persistent storage
│   ├── contracts/          # Smart contract engine
│   └── metrics/            # Prometheus metrics
├── web/                    # Web dashboard
└── docs/                   # Documentation
```

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/anekazek/simple-blockchain.git
cd simple-blockchain

# Run the core blockchain
go run main.go
```

### API Usage

Once the server is running, you can interact with the blockchain using HTTP requests:

#### Get the entire blockchain
```
GET http://localhost:8080/
```

#### Write data to the blockchain
```
POST http://localhost:8080/write
Content-Type: application/json

{
  "data": "Your data here"
}
```

## Architecture

The system follows a modular architecture with clear separation of concerns:

1. **Core Blockchain Layer**: Manages blocks, chain validation, and basic consensus
2. **Networking Layer**: Handles P2P communication and node synchronization
3. **Consensus Layer**: Implements and manages different consensus algorithms
4. **Storage Layer**: Provides persistent storage for the blockchain
5. **API Layer**: Exposes RESTful endpoints for external interaction
6. **Smart Contract Layer**: Executes and manages smart contracts
7. **Monitoring Layer**: Collects and exposes metrics for system health

## License

This project is licensed under the MIT License - see the LICENSE file for details.