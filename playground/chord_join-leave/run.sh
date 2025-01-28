#!/bin/bash

# Configuration
BASE_PORT=5000      # Starting port number
NODE_COUNT=10       # Number of nodes to start
HOSTNAME="localhost" # Hostname for nodes
EXECUTABLE="go run ."  # Path to your Chord executable (update if needed)
LOG_DIR="./logs"     # Directory for log files

# Ensure the log directory exists
mkdir -p "$LOG_DIR"

# Start the bootstrap node (first node)
echo "Starting bootstrap node on $HOSTNAME:$BASE_PORT"
$EXECUTABLE "$HOSTNAME:$BASE_PORT" > "$LOG_DIR/node_0.log" 2>&1 &
BOOTSTRAP_PID=$!
sleep 2 # Give the bootstrap node time to initialize

if ! kill -0 $BOOTSTRAP_PID >/dev/null 2>&1; then
    echo "Failed to start bootstrap node. Exiting."
    exit 1
fi

# Start additional nodes
for i in $(seq 1 $((NODE_COUNT - 1))); do
    PORT=$((BASE_PORT + i))
    echo "Starting node $i on $HOSTNAME:$PORT"
    $EXECUTABLE "$HOSTNAME:$PORT" > "$LOG_DIR/node_$i.log" 2>&1 &
    NODE_PID=$!
    sleep 2 # Give each node time to start

    if ! kill -0 $NODE_PID >/dev/null 2>&1; then
        echo "Failed to start node $i on $HOSTNAME:$PORT. Exiting."
        exit 1
    fi

    # Join the node to the ring via the bootstrap node
    echo "Joining node $i ($HOSTNAME:$PORT) to the ring via $HOSTNAME:$BASE_PORT"
    curl -X POST -d '' "$HOSTNAME:$BASE_PORT/join?nprime=$HOSTNAME:$PORT" -w "\n" || {
        echo "Failed to join node $i ($HOSTNAME:$PORT) to the ring. Exiting."
        exit 1
    }
done

echo "All $NODE_COUNT nodes have been started and joined the ring."

# Optional: Monitor logs or keep the script running
echo "Use 'tail -f $LOG_DIR/node_*.log' to monitor logs."
wait

