#!/bin/bash

# Configuration
BASE_PORT=9000      # Starting port number
NODE_COUNT=3       # Number of nodes to start
HOSTNAME="localhost" # Hostname for nodes
SERVER_DIR="../../"
EXECUTABLE="./owlmq_server"  # Path to your Chord executable (update if needed)
LOG_DIR="./logs"     # Directory for log files

# Function to check if a port is in use
is_port_in_use() {
  local port=$1
  if ss -ltn | grep -q ":$port"; then
    return 0  # Port is in use
  else
    return 1  # Port is free
  fi
}

# Function to generate a free port
generate_free_port() {
  while :; do
    # Randomly pick a port in the range 1024-65535
    local port=$((RANDOM + 1024))
    if ! is_port_in_use "$port"; then
      echo $port
      return 0
    fi
  done
}

# Ensure the log directory exists
mkdir -p "$LOG_DIR"

# Sicherstellen, dass das Log-Verzeichnis existiert
mkdir -p "$LOG_DIR"

# Wechsel ins Server-Verzeichnis und baue die ausführbare Datei
echo "Building the gRPC server..."
WD=$(pwd)
cd "$SERVER_DIR" || { echo "Failed to enter the grpc_server directory. Exiting."; exit 1; }

# Go build für das gRPC-Server-Programm
go build -o "$EXECUTABLE"
if [ $? -ne 0 ]; then
    echo "Go build failed. Exiting."
    exit 1
fi

# Verschiebe die ausführbare Datei in das Skriptverzeichnis
mv "$EXECUTABLE" "$WD/$EXECUTABLE"
cd $WD || { echo "Failed to return to the scripts directory. Exiting."; exit 1; }

echo "Build successful. Executable is now available at $(pwd)/$EXECUTABLE"


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
    PORT=$(generate_free_port)

    echo "Starting node $i on $HOSTNAME:$PORT"
    $EXECUTABLE "$HOSTNAME:$PORT" > "$LOG_DIR/node_$i.log" 2>&1 &
    NODE_PID=$!
    sleep 2 # Give each node time to start

    if ! kill -0 $NODE_PID >/dev/null 2>&1; then
        echo "Failed to start node $i on $HOSTNAME:$PORT. Exiting."
        exit 1
    fi

    # Join the node to the ring via gRPC (using Go gRPC client)
    echo "Joining node $i ($HOSTNAME:$PORT) to the ring via $HOSTNAME:$BASE_PORT"

    # gRPC client to send join request
    go run join_grpc_client.go "$HOSTNAME:$BASE_PORT" "$HOSTNAME:$PORT"

    if [ $? -ne 0 ]; then
        echo "Failed to join node $i ($HOSTNAME:$PORT) to the ring. Exiting."
        exit 1
    fi
done

echo "All $NODE_COUNT nodes have been started and joined the ring."

# Optional: Monitor logs or keep the script running
echo "Use 'tail -f $LOG_DIR/node_*.log' to monitor logs."
wait

