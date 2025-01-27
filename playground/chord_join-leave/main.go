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
	"strings"
	"time"
)

// Node stellt einen Knoten im DHT dar
type Node struct {
	ID          int               // Eindeutige ID des Knotens (basierend auf Hash)
	Address     string            // Adresse des Knotens (IP und Port)
	FingerTable []*Node           // Finger-Tabelle für Routing
	Successor   *Node             // Nachfolger im Chord-Ring
	Predecessor *Node             // Vorgänger im Chord-Ring
	Data        map[string]string // Schlüssel-Wert-Daten des Knotens
}

// Neue Knoteninstanz erstellen
func NewNode(id int, address string) *Node {
	return &Node{
		ID:          id,
		Address:     address,
		FingerTable: make([]*Node, 0), // Zu Beginn leer
		Data:        make(map[string]string),
	}
}

// Hash-Funktion für Schlüssel
func hashKey(key string) int {
	hash := sha1.New()
	hash.Write([]byte(key))
	hashBytes := hash.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes)
	hashInt, _ := hex.DecodeString(hashStr[:8])
	return int(hashInt[0])
}

// Initialisierung der Finger-Tabelle
func (n *Node) InitFingerTable(allNodes []*Node) {
	// Berechnung der Finger-Tabelle: Ein Finger zeigt auf den Knoten, der den nächstgelegenen größeren Hash-Wert hat.
	for i := 0; i < len(allNodes); i++ {
		finger := findSuccessor(allNodes, n.ID+int(1<<i)) // i ist bereits ein int
		n.FingerTable = append(n.FingerTable, finger)
	}
}

// Sucht den Knoten, der für einen bestimmten Hash-Wert zuständig ist
func findSuccessor(nodes []*Node, target int) *Node {
	for _, node := range nodes {
		if node.ID >= target {
			return node
		}
	}
	return nodes[0] // Zum Startknoten zurückkehren, wenn kein größerer Knoten gefunden wird
}

// HTTP-Handler für /storage/<key> (PUT und GET)
func (n *Node) handleStorage(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/storage/"):]

	switch r.Method {
	case "GET":
		// Hole den Wert für den Schlüssel
		value, ok := n.Data[key]
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(value))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}

	case "PUT":
		// Lese den Wert aus dem Body der Anfrage (kein JSON, nur einfacher Text)
		value := ""
		if r.Body != nil {
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body) // Liest den gesamten Body
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			value = string(body) // Umwandlung von []byte zu string
		}

		// Setze den Wert im DHT
		n.Data[key] = value
		w.WriteHeader(http.StatusOK)
	}
}

// HTTP-Handler für /network (Liste aller Knoten im Netzwerk)
func (n *Node) handleNetwork(w http.ResponseWriter, r *http.Request) {
	// Gibt eine Liste aller bekannten Knoten zurück
	json.NewEncoder(w).Encode(n.FingerTable)
}

// Starten des HTTP-Servers für einen Knoten
func startServer(n *Node) {
	http.HandleFunc("/storage/", n.handleStorage)
	http.HandleFunc("/network", n.handleNetwork)
	log.Fatal(http.ListenAndServe(n.Address, nil))
}

// Test-Put-Anfrage an einen Knoten senden
func testPut(key, value string, nodeAddress string) {
	start := time.Now()
	resp, err := http.Post(fmt.Sprintf("http://%s/storage/%s", nodeAddress, key), "text/plain", strings.NewReader(value))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("PUT %s took %v\n", key, time.Since(start))
	resp.Body.Close()
}

// Test-Get-Anfrage an einen Knoten senden
func testGet(key string, nodeAddress string) {
	start := time.Now()
	resp, err := http.Get(fmt.Sprintf("http://%s/storage/%s", nodeAddress, key))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("GET %s took %v\n", key, time.Since(start))
	resp.Body.Close()
}

// Hauptfunktion für das Starten eines Knotens und des Systems
func main() {
	// Überprüfe, ob die Adresse als Argument übergeben wurde
	if len(os.Args) < 2 {
		log.Fatal("Bitte eine Adresse im Format 'host:port' als Argument angeben")
	}
	// Das erste Argument ist die Adresse des Knotens (z.B. localhost:5000)
	address := os.Args[1]

	// Initialisiere einen Knoten mit der angegebenen Adresse
	node := NewNode(hashKey(address), address)

	// Initialisiere Finger-Tabelle für den Knoten (nur für einen Knoten in diesem Fall)
	// Für nur einen Knoten ist die Finger-Tabelle leer (da er kein Nachbar hat)
	node.InitFingerTable([]*Node{node})

	// Starte den Knoten-HTTP-Server
	go startServer(node)

	// Blockiere den Haupt-Thread, damit der Server weiterläuft
	select {}
}
