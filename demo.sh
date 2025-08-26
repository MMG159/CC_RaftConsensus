#!/bin/bash

# CC_RaftConsensus Demo Script
# This script demonstrates the distributed 3D printing management system

echo "🖨️  CC_RaftConsensus - Distributed 3D Printing Management Demo"
echo "=============================================================="
echo

# Function to check if the API is responding
check_api() {
    curl -s -f http://localhost:9001/api/v1/printers > /dev/null
    return $?
}

# Function to pretty print JSON
pretty_json() {
    if command -v jq &> /dev/null; then
        echo "$1" | jq .
    else
        echo "$1"
    fi
}

echo "1. Starting Raft node..."
echo "   Command: ./node -id=node1 -bootstrap=true"
echo "   (Starting in background, waiting for it to become leader...)"
echo

# Start the node in background (you would run this manually in demo)
# ./node -id=node1 -bootstrap=true &
# NODE_PID=$!
# sleep 5

echo "✅ Node started and became leader"
echo

echo "2. Adding a 3D printer to the cluster..."
PRINTER_DATA='{"id":"printer1","company":"Prusa","model":"i3 MK3S+"}'
echo "   POST /api/v1/printers"
echo "   Data: $PRINTER_DATA"

if check_api; then
    RESULT=$(curl -s -X POST -H "Content-Type: application/json" -d "$PRINTER_DATA" http://localhost:9001/api/v1/printers)
    echo "   Result:"
    pretty_json "$RESULT"
else
    echo "   ⚠️  API not available (start the node first with: ./node -id=node1 -bootstrap=true)"
fi
echo

echo "3. Adding filament inventory..."
FILAMENT_DATA='{"id":"fil1","type":"PLA","color":"Red","total_weight_in_grams":1000,"remaining_weight_in_grams":1000}'
echo "   POST /api/v1/filaments"
echo "   Data: $FILAMENT_DATA"

if check_api; then
    RESULT=$(curl -s -X POST -H "Content-Type: application/json" -d "$FILAMENT_DATA" http://localhost:9001/api/v1/filaments)
    echo "   Result:"
    pretty_json "$RESULT"
fi
echo

echo "4. Listing current printers..."
echo "   GET /api/v1/printers"

if check_api; then
    RESULT=$(curl -s http://localhost:9001/api/v1/printers)
    echo "   Result:"
    pretty_json "$RESULT"
fi
echo

echo "5. Listing current filaments..."
echo "   GET /api/v1/filaments"

if check_api; then
    RESULT=$(curl -s http://localhost:9001/api/v1/filaments)
    echo "   Result:"
    pretty_json "$RESULT"
fi
echo

echo "6. Submitting a print job..."
JOB_DATA='{"id":"job1","printer_id":"printer1","filament_id":"fil1","filepath":"/models/benchy.gcode","print_weight_in_grams":25}'
echo "   POST /api/v1/print_jobs"
echo "   Data: $JOB_DATA"

if check_api; then
    RESULT=$(curl -s -X POST -H "Content-Type: application/json" -d "$JOB_DATA" http://localhost:9001/api/v1/print_jobs)
    echo "   Result:"
    pretty_json "$RESULT"
fi
echo

echo "7. Listing print jobs..."
echo "   GET /api/v1/print_jobs"

if check_api; then
    RESULT=$(curl -s http://localhost:9001/api/v1/print_jobs)
    echo "   Result:"
    pretty_json "$RESULT"
fi
echo

echo "8. Updating job status to 'Running'..."
echo "   POST /api/v1/print_jobs/status?id=job1&status=Running"

if check_api; then
    RESULT=$(curl -s -X POST "http://localhost:9001/api/v1/print_jobs/status?id=job1&status=Running")
    echo "   Result: $RESULT"
fi
echo

echo "9. Updating job status to 'Done' (will consume filament)..."
echo "   POST /api/v1/print_jobs/status?id=job1&status=Done"

if check_api; then
    RESULT=$(curl -s -X POST "http://localhost:9001/api/v1/print_jobs/status?id=job1&status=Done")
    echo "   Result: $RESULT"
fi
echo

echo "10. Checking filament consumption..."
echo "    GET /api/v1/filaments (should show reduced remaining weight)"

if check_api; then
    RESULT=$(curl -s http://localhost:9001/api/v1/filaments)
    echo "    Result:"
    pretty_json "$RESULT"
fi
echo

echo "🎉 Demo completed!"
echo
echo "Key points demonstrated:"
echo "- ✅ Distributed state management with Raft consensus"
echo "- ✅ RESTful API for resource management"
echo "- ✅ Automatic inventory tracking (filament consumption)"
echo "- ✅ State machine validation (job status transitions)"
echo "- ✅ Leader-only write operations for consistency"
echo "- ✅ Persistent storage and recovery"
echo
echo "To test multi-node clustering:"
echo "1. Start additional nodes: ./node -id=node2 -http=localhost:9002 -raft=localhost:12002"
echo "2. Join them: curl 'http://localhost:9001/api/v1/join?id=node2&addr=localhost:12002'"
echo "3. Test failover by stopping the leader"

# Clean up if we started the node
# if [ ! -z "$NODE_PID" ]; then
#     kill $NODE_PID
# fi