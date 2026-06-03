# CC_RaftConsensus (raft3d)

A minimal **Raft-backed service in Go** (HashiCorp Raft) that replicates application state across nodes and exposes a simple HTTP API for managing a small “3D printing shop” domain:

- **Printers**
- **Filaments** (with remaining weight tracking)
- **Print jobs** (with a small status state-machine)

This repository is primarily a learning/demo implementation of consensus and state replication using the [HashiCorp Raft](https://github.com/hashicorp/raft) library.

## What’s in this repo

- `cmd/node` — runnable node binary (HTTP API + Raft)
- `raftcore` — Raft initialization + FSM implementation
- `api` — HTTP handlers that read/write state via the Raft log
- `models` — request/response types (Printer / Filament / PrintJob)
- `raft-data-node*` — example data directories (logs/snapshots) created by nodes

## How it works (high level)

- Each node runs:
  - an **HTTP server** (for client requests)
  - a **Raft transport** listener (for peer-to-peer consensus traffic)
- **Writes** (create printer/filament/job, update job status, join node) are only accepted by the **leader**.
- Write requests are encoded as a `raftcore.FSMCommand` and appended to the Raft log.
- The FSM (`raftcore/FSM`) applies committed commands to an in-memory state:
  - `Printers map[string]Printer`
  - `Filaments map[string]Filament`
  - `PrintJobs map[string]PrintJob`
- The FSM supports **snapshots** and **restore**.

## Requirements

- Go **1.24.2** (see `go.mod`)

## Build

```bash
go build ./...
```

## Run a 3-node cluster locally

Open **three terminals**.

### 1) Start node1 (bootstrap leader)

```bash
go run ./cmd/node \
  -id node1 \
  -http localhost:9001 \
  -raft localhost:12001 \
  -data raft-data-node1 \
  -bootstrap
```

### 2) Start node2

```bash
go run ./cmd/node \
  -id node2 \
  -http localhost:9002 \
  -raft localhost:12002 \
  -data raft-data-node2
```

### 3) Start node3

```bash
go run ./cmd/node \
  -id node3 \
  -http localhost:9003 \
  -raft localhost:12003 \
  -data raft-data-node3
```

### 4) Join node2 and node3 to the cluster (call node1)

```bash
curl -X POST "http://localhost:9001/api/v1/join?id=node2&addr=localhost:12002"
curl -X POST "http://localhost:9001/api/v1/join?id=node3&addr=localhost:12003"
```

Notes:
- Joining must be done against the **leader**.
- `addr` is the node’s **Raft bind address** (not the HTTP address).

## API

Base path: `http://<host>:<httpPort>/api/v1`

### Printers

- `GET /printers` — list printers
- `POST /printers` — create printer (leader only)

Example:

```bash
curl -s -X POST http://localhost:9001/api/v1/printers \
  -H 'Content-Type: application/json' \
  -d '{"id":"p1","company":"Prusa","model":"MK4"}'

curl -s http://localhost:9001/api/v1/printers
```

### Filaments

- `GET /filaments` — list filaments
- `POST /filaments` — create filament (leader only)

Example:

```bash
curl -s -X POST http://localhost:9001/api/v1/filaments \
  -H 'Content-Type: application/json' \
  -d '{"id":"f1","type":"PLA","color":"black","total_weight_in_grams":1000,"remaining_weight_in_grams":1000}'

curl -s http://localhost:9001/api/v1/filaments
```

### Print jobs

- `GET /print_jobs` — list print jobs
- `POST /print_jobs` — create print job (leader only)
  - Validates that `printer_id` and `filament_id` exist.
  - Initial status is set to `Queued` when the FSM applies it.

Example:

```bash
curl -s -X POST http://localhost:9001/api/v1/print_jobs \
  -H 'Content-Type: application/json' \
  -d '{"id":"j1","printer_id":"p1","filament_id":"f1","filepath":"/prints/calibration_cube.gcode","print_weight_in_grams":50}'

curl -s http://localhost:9001/api/v1/print_jobs
```

### Update job status

- `POST /print_jobs/status?id=<jobId>&status=<Queued|Running|Cancelled|Done>` (leader only)

Allowed transitions (enforced by the FSM):

- `Queued -> Running`
- `Running -> Done`
- `Queued|Running -> Cancelled`

When a job transitions to `Done`, the filament’s `remaining_weight_in_grams` is reduced by `print_weight_in_grams`.

Example:

```bash
# Start printing
curl -s -X POST "http://localhost:9001/api/v1/print_jobs/status?id=j1&status=Running"

# Finish printing (also decrements remaining filament weight)
curl -s -X POST "http://localhost:9001/api/v1/print_jobs/status?id=j1&status=Done"
```

## Important behavior / limitations

- This implementation keeps the FSM state **in memory** and relies on Raft logs + snapshots for durability.
- Leader checks are simple (`rn.IsLeader()`); non-leader writes return an error.
  - A production system would usually **redirect clients** to the leader.
- Authentication, authorization, and TLS are not implemented.
- The HTTP routing uses `net/http` with manual method checks.

## Development

Run tests (if/when added):

```bash
go test ./...
```

## License

No license file is currently present in this repository. If you intend others to use it, consider adding a `LICENSE`.
