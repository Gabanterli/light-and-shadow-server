# Light and Shadow MMORPG — Backend Bootstrap Setup & Execution Guide

This is the production-ready Go backend bootstrap for the **Light and Shadow** MMORPG. It implements a clean, distributed architecture separating the Gateway, Auth, and World services.

## Distributed Architecture

```
                       +-----------------------+
                       |   Godot Client (C#)   |
                       +-----------------------+
                                   |
                                   v (TCP Binary Protocol, Little Endian)
                       +-----------------------+
                       |     Gateway Server    |
                       +-----------------------+
                        /                     \
                       / (Internal HTTP RPC)   \ (Internal Tick Loop)
                      v                         v
            +-------------------+     +-------------------+
            |    Auth Server    |     |    World Server   |
            +-------------------+     +-------------------+
                      \                         /
                       \                       /
                        v                     v
               +-----------------+   +-----------------+
               | PostgreSQL Pool |   |   Redis Cache   |
               +-----------------+   +-----------------+
```

---

## Folder Structure

```
backend/
├── cmd/
│   ├── auth/
│   │   └── main.go       # Auth Server (DB verification, session tokens)
│   ├── gateway/
│   │   └── main.go       # Gateway Server (Entry TCP socket handler, dispatcher, heartbeats)
│   └── world/
│       └── main.go       # World Server (InGame spatial entity sync, 20Hz gameplay game loop)
├── config/
│   └── config.go         # Strongly typed, environment-based configuration loader
├── pkg/
│   ├── db/
│   │   ├── postgres.go   # Thread-safe PostgreSQL Connection Pool manager
│   │   └── redis.go      # Resilient Redis Client initializer and close handlers
│   ├── lifecycle/
│   │   └── lifecycle.go  # Unix-signal listening Graceful Shutdown Lifecycle Manager
│   ├── logger/
│   │   └── logger.go     # Structured, blazing-fast JSON logging handler using 'log/slog'
│   └── protocol/
│       └── protocol.go   # Core Little-Endian 8-byte header Packet Parser
├── docker-compose.yml    # Full-stack environment orchestration file
├── Dockerfile.auth       # Multi-stage optimized scratch Docker container for Auth
├── Dockerfile.gateway    # Multi-stage optimized scratch Docker container for Gateway
├── Dockerfile.world      # Multi-stage optimized scratch Docker container for World
├── go.mod                # Go module file specifying minimal external requirements
└── README.md             # This comprehensive execution manual
```

---

## Prerequisites

- **Go 1.21+** installed locally
- **Docker & Docker Compose** installed
- **PostgreSQL** and **Redis** (if running bare-metal/locally)

---

## How to Execute the Backend

### Method A: Orchestrated with Docker Compose (Recommended)

To launch the complete distributed backend (PostgreSQL, Redis, Auth, World, and Gateway servers) with a single command, run the following from the `backend/` directory:

```bash
docker-compose up --build
```

This will automatically build all Go binaries into lightweight alpine microservices, establish private networks, and initialize ports.

- **Gateway TCP Server**: listening on port `8080` (for game clients)
- **Gateway HTTP Health**: `http://localhost:9080/health`
- **Auth Server Health**: `http://localhost:8081/health`
- **World Server Health**: `http://localhost:8082/health`

---

### Method B: Manual Bare-Metal Execution

If you prefer compiling and running the Go binaries locally on your development machine, configure your env vars and run:

1. **Start PostgreSQL & Redis** services on your local computer.
2. **Configure Environment Variables** (optional, fallback defaults exist):
   ```bash
   export GATEWAY_PORT=8080
   export AUTH_PORT=8081
   export WORLD_PORT=8082
   export POSTGRES_DSN="postgres://postgres:postgrespassword@localhost:5432/light_and_shadow?sslmode=disable"
   export REDIS_ADDR="localhost:6379"
   export REDIS_PASSWORD=""
   ```
3. **Execute Services**:
   Open three terminal tabs inside the `backend` folder and run:
   
   - **Tab 1: Auth Server**
     ```bash
     go run cmd/auth/main.go
     ```
   - **Tab 2: World Server**
     ```bash
     go run cmd/world/main.go
     ```
   - **Tab 3: Gateway Server**
     ```bash
     go run cmd/gateway/main.go
     ```

---

## Packet Protocol Design

Each binary TCP packet has an **8-byte header** parsed using Little-Endian order, followed by an optional payload:

| Offset | Data Type | Field | Description |
|---|---|---|---|
| `0x00` | `uint16` | `Size` | Total packet size (Header + Payload) |
| `0x02` | `uint16` | `Opcode` | Operation code mapping the request/response |
| `0x04` | `uint32` | `Sequence` | Atomic sequential identifier to track packets |
| `0x08` | `[]byte` | `Payload` | Variable binary chunk carrying actual payload |
