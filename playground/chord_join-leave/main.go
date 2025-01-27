package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// Node represents a node in the distributed hash table (DHT)
type Node struct {
	ID          int               // Unique ID of the node (based on hash)
	Address     string            // Address (host:port) of the node
	FingerTable []*Node           // Finger table for routing
	Successor   *Node             // Successor node in the Chord ring
	Predecessor *Node             // Predecessor node in the Chord ring
	Data        map[string]string // Key-value store for the node
	mu          sync.Mutex        // Mutex for thread-safe operations
}

// NewNode creates a new node with a given ID and address
func NewNode(id int, address string) *Node {
	return &Node{
		ID:          id,
		Address:     address,
		FingerTable: make([]*Node, 0), // Initially empty
		Data:        make(map[string]string),
	}
}

// hashKey hashes the key to an integer ID
func hashKey(key string) int {
	hash := sha1.New()
	hash.Write([]byte(key))
	hashBytes := hash.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes)
	hashInt, _ := hex.DecodeString(hashStr[:8])
	return int(hashInt[0])
}

// Initialize the Finger Table for the node
func (n *Node) InitFingerTable(allNodes []*Node) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Initialize the finger table by assigning nodes based on the ring
	for _, node := range allNodes {
		// Do not add the node itself to the finger table
		if node.ID != n.ID && !contains(n.FingerTable, node) {
			n.FingerTable = append(n.FingerTable, node)
		}
	}

	// Update successor and predecessor if there are nodes in the finger table
	if len(n.FingerTable) > 0 {
		n.Successor = n.FingerTable[0]
		n.Predecessor = n.FingerTable[len(n.FingerTable)-1]
	}
}

// contains checks if a node is already in the list of nodes (to avoid duplicates)
func contains(nodes []*Node, node *Node) bool {
	for _, n := range nodes {
		if n.ID == node.ID {
			return true
		}
	}
	return false
}

// handleStorage handles PUT and GET requests for the storage (key-value store)
func (n *Node) handleStorage(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/storage/"):]

	switch r.Method {
	case "GET":
		// Get the value for the key
		value, ok := n.Data[key]
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(value))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}

	case "PUT":
		// Read the value from the request body
		value := ""
		if r.Body != nil {
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body) // Use io.ReadAll instead of ioutil.ReadAll
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			value = string(body) // Convert []byte to string
		}

		// Store the key-value pair
		n.Data[key] = value
		w.WriteHeader(http.StatusOK)
	}
}

// handleNetwork handles the /network endpoint, returning the known nodes
func (n *Node) handleNetwork(w http.ResponseWriter, r *http.Request) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Create a list of nodes to return
	network := []map[string]string{
		{"ID": fmt.Sprintf("%d", n.ID), "Address": n.Address},
	}

	// Add the successor and predecessor to the list of known nodes
	if n.Successor != nil {
		network = append(network, map[string]string{"ID": fmt.Sprintf("%d", n.Successor.ID), "Address": n.Successor.Address})
	}
	if n.Predecessor != nil {
		network = append(network, map[string]string{"ID": fmt.Sprintf("%d", n.Predecessor.ID), "Address": n.Predecessor.Address})
	}

	// Add all known nodes in the finger table
	for _, finger := range n.FingerTable {
		network = append(network, map[string]string{"ID": fmt.Sprintf("%d", finger.ID), "Address": finger.Address})
	}

	// Return the list as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(network)
}

// joinNetwork adds a new node to the network by adding it to the finger tables
func (n *Node) joinNetwork(newNode *Node) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Add the new node to the finger table only if not already present
	if !contains(n.FingerTable, newNode) {
		n.FingerTable = append(n.FingerTable, newNode)
	}

	// Update successor and predecessor relationships
	newNode.Successor = n
	n.Predecessor = newNode
}

// leaveNetwork removes the current node from the network
func (n *Node) leaveNetwork() {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Remove the node from the finger tables of other nodes
	for _, node := range n.FingerTable {
		node.mu.Lock()
		defer node.mu.Unlock()

		// Remove the current node from the finger table of other nodes
		var updatedFingerTable []*Node
		for _, finger := range node.FingerTable {
			if finger.ID != n.ID {
				updatedFingerTable = append(updatedFingerTable, finger)
			}
		}
		node.FingerTable = updatedFingerTable
	}

	// Update successor and predecessor
	n.Successor = nil
	n.Predecessor = nil
}

// startServer starts the HTTP server for a given node
func startServer(n *Node) {
	http.HandleFunc("/storage/", n.handleStorage)
	http.HandleFunc("/network", n.handleNetwork)

	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		// Handle join requests
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address is required", http.StatusBadRequest)
			return
		}

		// New node joins the network
		newNode := NewNode(hashKey(address), address)
		n.joinNetwork(newNode)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
		// Handle leave requests
		n.leaveNetwork()
		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(n.Address, nil))
}

// main function to run the DHT node and start the server
func main() {
	// Check if address is provided as a command line argument
	if len(os.Args) < 2 {
		log.Fatal("Please provide an address in the format 'host:port' as an argument")
	}

	// The first argument is the node's address (e.g., localhost:5000)
	address := os.Args[1]

	// Create a new node with the given address
	node := NewNode(hashKey(address), address)

	// Start the HTTP server for the node
	go startServer(node)

	// Block the main thread to keep the server running
	select {}
}
