# MockServiceInGo

MockServiceInGo is an API mocking service for development and testing.  
It lets you register mock routes at runtime and return HTTP, SSE, or WebSocket responses without changing application code.

## Features

- Dynamic mock registration via HTTP API
- File-based rule persistence with periodic autosave
- Config upload endpoint to load many rules at once
- Distributed deployment mode using Raft for rule replication
- Optional React UI for managing routes (`frontend/MockService`)

## Current Status

- Distributed deployment is implemented (Raft-based replication across nodes).
- Built-in tenant-isolated storage is **not implemented yet** (roadmap item).
  - Current workaround: use tenant prefixes in paths (for example `/t/acme/orders`) or run one service instance per tenant.

## Repository Layout

- `backend/`: Go service and Raft implementation
- `frontend/MockService/`: optional React UI
- `backend/integrate_test/`: example standalone and cluster configs/scripts

## Prerequisites

- Go `1.25.x` (module in `backend/go.mod`)
- Optional for UI: Node.js + npm

## Quick Start (Backend, Standalone)

From project root:

```bash
cd backend
go run ./cmd/starter -config ./integrate_test/start-standalone.json
```

Service starts on `:9900` using the sample standalone config.

Health check:

```bash
curl -s http://127.0.0.1:9900/v1/health
```

## API Overview

Base API prefix: `/v1`

- `GET /v1/health`
- `POST /v1/__mock`
- `POST /v1/__mock/upload` (multipart form file field name must be `config.json`)
- `GET /v1/__mock`
- `DELETE /v1/__mock/all`
- `DELETE /v1/__mock/:method?path=/your/path`

All unmatched routes are handled by the mock matcher (`NoRoute`), so requests to registered paths return configured mock responses.

## Register Mock Examples

### 1) HTTP mock

```bash
curl -X POST http://127.0.0.1:9900/v1/__mock \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/orders",
    "status": 200,
    "responseHeaders": {
      "Content-Type": "application/json",
      "X-Mock": "true"
    },
    "responseBody": "{\"items\":[{\"id\":1}]}"
  }'
```

Test it:

```bash
curl -i http://127.0.0.1:9900/api/orders
```

### 2) SSE mock

```bash
curl -X POST http://127.0.0.1:9900/v1/__mock \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/stream/events",
    "responseType": "sse",
    "sseEvents": [
      { "event": "tick", "data": "one", "delay": 1000000000 },
      { "event": "tick", "data": "two", "delay": 1000000000 }
    ]
  }'
```

Notes:
- `delay` uses Go duration JSON encoding (`time.Duration`, nanoseconds).  
  `1000000000` = `1s`.

### 3) WebSocket mock

```bash
curl -X POST http://127.0.0.1:9900/v1/__mock \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/ws/demo",
    "responseType": "websocket",
    "websocketMessages": [
      { "message": "hello", "delay": 500000000, "type": "text" },
      { "message": "world", "delay": 500000000, "type": "text" }
    ]
  }'
```

Notes:
- For WebSocket mocks, server-side logic enforces `GET` for upgrade requests.

## Upload Rules From File

Upload a JSON file containing an array of rules:

```bash
curl -X POST http://127.0.0.1:9900/v1/__mock/upload \
  -F "config.json=@./config_upload_test.json"
```

## Persistence Behavior

- Rules are loaded from `rulesFile` at startup (if file exists).
- Rules are auto-saved approximately every 2 seconds.
- In distributed mode, rule mutations are replicated through Raft.

## Distributed Deployment (Raft Cluster)

Sample 3-node configs:

- `backend/integrate_test/start1.json` (service `:9000`, raft `127.0.0.1:8081`)
- `backend/integrate_test/start2.json` (service `:9001`, raft `127.0.0.1:8082`)
- `backend/integrate_test/start3.json` (service `:9002`, raft `127.0.0.1:8083`)

Start cluster:

```bash
cd backend/integrate_test
./start_cluster.sh
```

Stop cluster:

```bash
cd backend/integrate_test
./stop.sh
```

Manual start option:

```bash
cd backend
go run ./cmd/starter -config ./integrate_test/start1.json
go run ./cmd/starter -config ./integrate_test/start2.json
go run ./cmd/starter -config ./integrate_test/start3.json
```

Cluster behavior notes:

- Writes must be accepted by the current leader to replicate.
- Followers may reject writes during leadership mismatch/election windows.
- After leader election settles, successful writes propagate to all nodes.

## Optional Frontend UI

From project root:

```bash
cd frontend/MockService
npm install
npm run dev
```

Other scripts:

```bash
npm run build
npm run test
npm run lint
```

UI expects backend API to be running.

## Development Commands

Backend:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend/MockService
npm run test
npm run lint
```

## Troubleshooting

- `config file is required`: start `starter` with `-config`.
- Port bind errors: update `serviceAddr`/`raft.address` in config JSON files.
- Cluster write conflicts: retry after election; ensure request targets leader node.
- Missing responses: verify request method + path exactly match registered rule.
