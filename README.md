# CC_RaftConsensus

A distributed 3D printing management system built with the Raft consensus algorithm in Go.

## Quick Start

1. **Build the application:**
   ```bash
   go build ./cmd/node
   ```

2. **Start a single node:**
   ```bash
   ./node -id=node1 -bootstrap=true
   ```

3. **Run the demo:**
   ```bash
   ./demo.sh
   ```

## What This Project Demonstrates

This project showcases a real-world application of distributed systems concepts:

- **Raft Consensus Algorithm** for distributed coordination
- **Leader-based write operations** ensuring data consistency
- **Automatic failover** and leader election
- **Persistent storage** with recovery capabilities
- **RESTful API** for resource management
- **State machine** design for business logic

## Use Case

Manages a distributed 3D printing farm with:
- Multiple printers across different locations
- Filament inventory tracking
- Print job queue coordination
- Automatic consumption calculation

## Project Structure

```
├── api/            # HTTP API handlers
├── cmd/node/       # Main application
├── models/         # Data structures
├── raftcore/       # Raft consensus implementation
├── demo.sh         # Demonstration script
└── PROJECT_OVERVIEW.md  # Detailed technical explanation
```

## For Interview Context

See `PROJECT_OVERVIEW.md` for a comprehensive explanation of the project's technical concepts, architecture decisions, and discussion points suitable for technical interviews.

## API Examples

```bash
# Add a printer
curl -X POST -H "Content-Type: application/json" \
  -d '{"id":"printer1","company":"Prusa","model":"i3 MK3S+"}' \
  http://localhost:9001/api/v1/printers

# Add filament
curl -X POST -H "Content-Type: application/json" \
  -d '{"id":"fil1","type":"PLA","color":"Red","total_weight_in_grams":1000,"remaining_weight_in_grams":1000}' \
  http://localhost:9001/api/v1/filaments

# Submit print job
curl -X POST -H "Content-Type: application/json" \
  -d '{"id":"job1","printer_id":"printer1","filament_id":"fil1","filepath":"/models/test.gcode","print_weight_in_grams":50}' \
  http://localhost:9001/api/v1/print_jobs
```

## Multi-Node Setup

```bash
# Start bootstrap node
./node -id=node1 -http=localhost:9001 -raft=localhost:12001 -bootstrap=true

# Start additional nodes
./node -id=node2 -http=localhost:9002 -raft=localhost:12002
./node -id=node3 -http=localhost:9003 -raft=localhost:12003

# Join nodes to cluster
curl "http://localhost:9001/api/v1/join?id=node2&addr=localhost:12002"
curl "http://localhost:9001/api/v1/join?id=node3&addr=localhost:12003"
```

## Dependencies

- Go 1.24+
- [HashiCorp Raft](https://github.com/hashicorp/raft)
- [BoltDB](https://github.com/boltdb/bolt) for persistence