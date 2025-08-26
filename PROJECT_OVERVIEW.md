# CC_RaftConsensus - Distributed 3D Printing Management System

## Project Overview for Interview Context

This project is a **distributed 3D printing management system** built using the **Raft consensus algorithm** in Go. It demonstrates the application of distributed systems concepts to solve real-world coordination problems in a 3D printing farm or makerspace environment.

## What Problem Does It Solve?

In a large-scale 3D printing environment (like a manufacturing facility or university lab), you need to:
- Manage multiple 3D printers across different locations
- Track filament inventory and consumption
- Queue and coordinate print jobs efficiently
- Ensure data consistency across all management nodes
- Provide high availability even if some management servers fail

## Key Technical Concepts Demonstrated

### 1. **Raft Consensus Algorithm**
- **Purpose**: Ensures all nodes in the cluster agree on the current state
- **Use Case**: When a print job is submitted, all nodes must agree it exists and track its status consistently
- **Implementation**: Uses HashiCorp's Raft library for leader election, log replication, and fault tolerance

### 2. **Finite State Machine (FSM)**
- **Purpose**: Manages the application state (printers, filaments, print jobs)
- **Design**: Thread-safe operations with mutex locks
- **Operations**: CREATE, UPDATE, and DELETE operations on resources

### 3. **Distributed Architecture**
- **Leader Node**: Handles write operations (creating jobs, updating status)
- **Follower Nodes**: Replicate data and can handle read operations
- **Fault Tolerance**: If leader fails, followers elect a new leader automatically

## System Components

### 1. **Data Models** (`models/types.go`)
```go
type Printer struct {
    ID      string `json:"id"`
    Company string `json:"company"`
    Model   string `json:"model"`
}

type Filament struct {
    ID                     string `json:"id"`
    Type                   string `json:"type"`    // PLA, PETG, ABS, TPU
    Color                  string `json:"color"`
    TotalWeightInGrams     int    `json:"total_weight_in_grams"`
    RemainingWeightInGrams int    `json:"remaining_weight_in_grams"`
}

type PrintJob struct {
    ID                 string `json:"id"`
    PrinterID          string `json:"printer_id"`
    FilamentID         string `json:"filament_id"`
    Filepath           string `json:"filepath"`
    PrintWeightInGrams int    `json:"print_weight_in_grams"`
    Status             string `json:"status"` // Queued, Running, Cancelled, Done
}
```

### 2. **Raft Core** (`raftcore/`)
- **`raft.go`**: Raft node initialization, configuration, and peer management
- **`fsm.go`**: Finite State Machine implementing business logic

### 3. **HTTP API** (`api/handlers.go`)
- RESTful endpoints for managing printers, filaments, and print jobs
- Leader-only write operations ensure consistency
- Read operations can be served by any node

### 4. **Node Application** (`cmd/node/main.go`)
- Command-line application that starts a Raft node
- Configurable ports for HTTP API and Raft communication
- Bootstrap capability for cluster initialization

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/printers` | List all printers |
| POST | `/api/v1/printers` | Add a new printer |
| GET | `/api/v1/filaments` | List all filaments |
| POST | `/api/v1/filaments` | Add a new filament |
| GET | `/api/v1/print_jobs` | List all print jobs |
| POST | `/api/v1/print_jobs` | Submit a new print job |
| POST | `/api/v1/print_jobs/status` | Update job status |
| POST | `/api/v1/join` | Add a node to the cluster |

## How to Run

### Single Node (Development)
```bash
./node -id=node1 -bootstrap=true
```

### Multi-Node Cluster
```bash
# Node 1 (Bootstrap)
./node -id=node1 -http=localhost:9001 -raft=localhost:12001 -bootstrap=true

# Node 2
./node -id=node2 -http=localhost:9002 -raft=localhost:12002

# Node 3
./node -id=node3 -http=localhost:9003 -raft=localhost:12003

# Join nodes to cluster
curl "http://localhost:9001/api/v1/join?id=node2&addr=localhost:12002"
curl "http://localhost:9001/api/v1/join?id=node3&addr=localhost:12003"
```

## Example Usage

### 1. Add a Printer
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"id":"printer1","company":"Prusa","model":"i3 MK3S+"}' \
  http://localhost:9001/api/v1/printers
```

### 2. Add Filament
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"id":"fil1","type":"PLA","color":"Red","total_weight_in_grams":1000,"remaining_weight_in_grams":1000}' \
  http://localhost:9001/api/v1/filaments
```

### 3. Submit Print Job
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"id":"job1","printer_id":"printer1","filament_id":"fil1","filepath":"/models/test.gcode","print_weight_in_grams":50}' \
  http://localhost:9001/api/v1/print_jobs
```

### 4. Update Job Status
```bash
curl -X POST "http://localhost:9001/api/v1/print_jobs/status?id=job1&status=Running"
curl -X POST "http://localhost:9001/api/v1/print_jobs/status?id=job1&status=Done"
```

## Advanced Features Demonstrated

### 1. **Automatic Filament Consumption Tracking**
When a print job is marked as "Done", the system automatically:
- Deducts the used filament weight from the filament's remaining weight
- Updates the filament inventory in real-time

### 2. **State Validation**
- Print jobs can only be created if both printer and filament exist
- Status transitions follow valid state machine rules (Queued → Running → Done/Cancelled)

### 3. **Persistence and Recovery**
- Uses BoltDB for persistent storage of Raft logs
- Supports snapshots for efficient recovery
- Data survives node restarts and failures

### 4. **Leader-Only Writes**
- Only the Raft leader accepts write operations
- Ensures linearizable consistency across the cluster
- Automatic redirection to leader in a proper production setup

## Interview Discussion Points

### 1. **Why Raft over other consensus algorithms?**
- Raft is easier to understand than Paxos
- Strong leader model simplifies client interactions
- Excellent tooling and library support in Go ecosystem

### 2. **Scalability Considerations**
- Current design handles moderate-scale 3D printing operations
- For larger scale, could implement read replicas
- Could partition data by facility/location for horizontal scaling

### 3. **Real-World Production Considerations**
- Add authentication and authorization
- Implement proper error handling and retry logic
- Add metrics and monitoring
- Implement graceful shutdown procedures
- Add data validation and sanitization

### 4. **Alternative Approaches**
- Could use event sourcing with Apache Kafka
- Could use traditional database with proper transaction handling
- Could use blockchain for immutable audit trail

### 5. **Testing Strategy**
- Unit tests for business logic
- Integration tests for API endpoints
- Chaos engineering tests for fault tolerance
- Load testing for performance characteristics

## Technical Achievements

1. **Distributed Systems**: Implemented a working Raft cluster
2. **Concurrency**: Thread-safe FSM with proper locking
3. **REST API Design**: Clean, RESTful interface
4. **Data Persistence**: Durable storage with recovery capabilities
5. **Fault Tolerance**: Automatic leader election and failover
6. **Real-World Application**: Solves actual coordination problems in manufacturing

This project demonstrates understanding of distributed systems concepts, consensus algorithms, and their practical application to real-world coordination problems in industrial settings.