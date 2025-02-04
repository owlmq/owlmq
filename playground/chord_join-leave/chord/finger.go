package chord

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/owlmq/owlmq/playground/utils"
)

var (
	once_fingertable     sync.Once
	instance_fingertable *FingerTable
)

func GetFingerTable(ctx *context.Context) *FingerTable {
	once_fingertable.Do(func() {
		instance_fingertable = &FingerTable{
			ctx:            ctx,
			Entries:        []FingerEntry{},
			successorStack: stack{},
		}
	})

	return instance_fingertable
}

type FingerTable struct {
	ctx            *context.Context
	Entries        []FingerEntry
	successorStack stack
	mu             sync.Mutex
}

type FingerEntry struct {
	Hash     string
	NodeName string
}

// HexToInt converts a hex string into an integer.
func HexToInt(hexStr string) int {
	// Decode the hex string into bytes
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic("Invalid hex string")
	}

	// Convert bytes into a big integer
	bigInt := new(big.Int)
	bigInt.SetBytes(bytes)

	// Convert to int (ensure it's within range)
	return int(bigInt.Int64())
}

func (f *FingerTable) GenerateFingerTable(ctx *context.Context, nodeID string, hashSpaceBits int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	m := hashSpaceBits
	f.Entries = make([]FingerEntry, m) // Allocate space for m fingers

	nodeInt := HexToInt(nodeID) // Convert node ID from hex to int

	for i := 0; i < m; i++ {
		// Calculate start of the finger interval: (nodeID + 2^i) % 2^m
		start := (nodeInt + int(math.Pow(2, float64(i)))) % int(math.Pow(2, float64(m)))
		startHash := utils.GenerateSHA1(fmt.Sprintf("%d", start))

		// Find the node responsible for the interval start
		successor := f.findSuccessor(ctx, startHash)

		// Add the entry to the finger table
		f.Entries[i] = FingerEntry{
			Hash:     startHash,
			NodeName: successor,
		}
	}
}

// findSuccessor is a helper to find the node responsible for a given hash.
func (f *FingerTable) findSuccessor(ctx *context.Context, key string) string {
	// Logic to find the successor node
	hostname := (*ctx).Value("hostname").(string)
	//successor := (*ctx).Value("successor").(string)

	// If key is in this node's range, return self
	if utils.IsKeyInRange(key, hostname, (*ctx).Value("predecessor").(string)) {
		return hostname
	}

	// Otherwise, forward the request to the next node
	fe := f.GetOne(key)
	return fe.NodeName
}

// UpdateFingerTable updates entries based on new node join/leave events.
func (f *FingerTable) UpdateFingerTable(ctx *context.Context) {
	// Re-run GenerateFingerTable to refresh entries
	nodeID := utils.GenerateSHA1((*ctx).Value("hostname").(string))
	f.GenerateFingerTable(ctx, nodeID, 160) // Assuming 160-bit hash space
}

// GetOne fetches the closest preceding node for a given key.
func (f *FingerTable) GetOne(key string) *FingerEntry {
	hk := utils.GenerateSHA1(key)

	for i := len(f.Entries) - 1; i >= 0; i-- {
		if utils.GenerateSHA1(f.Entries[i].NodeName) < hk {
			return &f.Entries[i]
		}
	}
	return &f.Entries[len(f.Entries)-1]
}

func (f *FingerTable) GetAll() ([]FingerEntry, error) {
	if len(f.Entries) == 0 {
		return nil, errors.New("list is empty")
	}
	return f.Entries, nil
}

func (f *FingerTable) stabilize(ctx *context.Context) {
	hostname := (*ctx).Value("hostname").(string)
	successor := (*ctx).Value("successor").(string)
	if successor == "" {
		return
	}

	//get predecessor of my successor
	requestURL := fmt.Sprintf("http://%s/predecessor", successor)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Printf("error creating http get request: %s\n", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		//log.Printf("error executing http get request: %s\n", err)
		//successor crashed -> find my new successor
		return
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error making http request: %s\n", err)
	}

	//fmt.Printf("-%v-\n", utils.IsKeyInRange(pertantialSuc, utils.GenerateSHA1(successor), utils.GenerateSHA1(hostname)))
	if string(resBody) != "" && utils.IsKeyInRange(string(resBody), successor, hostname) {
		(*ctx) = context.WithValue((*ctx), "successor", string(resBody))
	}
	f.notify(ctx)

}

func (f *FingerTable) notify(ctx *context.Context) {
	//notify my successor that i am his predecessor
	requestURL := fmt.Sprintf("http://%s/predecessor", (*ctx).Value("successor").(string))

	req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer([]byte((*ctx).Value("hostname").(string))))
	if err != nil {
		log.Printf("error creating http get request: %s\n", err)
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("error executing http get request: %s\n", err)
		return
	}
}

func (f *FingerTable) fixFingers(ctx *context.Context) {
	f.Entries = []FingerEntry{
		{
			Hash:     utils.GenerateSHA1((*ctx).Value("successor").(string)),
			NodeName: (*ctx).Value("successor").(string),
		},
	}
}

func (f *FingerTable) checkPredecessorAlive(ctx *context.Context) {
	predecessor := (*ctx).Value("predecessor").(string)
	if predecessor == "" {
		return
	}

	if checkIfNodeIsAlive(predecessor) != true {
		//fmt.Println("predecessor crashed")
		(*ctx) = context.WithValue((*ctx), "predecessor", "")
	}
}

func (f *FingerTable) fixSuccessorList(ctx *context.Context) {
	successor := (*ctx).Value("successor").(string)
	if successor == "" {
		return
	}

	//test if successor is up
	if checkIfNodeIsAlive(successor) != true {
		//fmt.Println("successor crashed")
		//get next successor from the successorStack
		nextSuc, err := f.successorStack.Pop()
		if err != nil {
			//stack is empty so we dont know a new successor which means we are alone
			(*ctx) = context.WithValue((*ctx), "successor", (*ctx).Value("hostname").(string))
		} else {
			//fmt.Println("new successor is ", nextSuc)
			(*ctx) = context.WithValue((*ctx), "successor", nextSuc)
		}

	}

	//generate successor list
	f.successorStack.Free()

	nextToAsk := (*ctx).Value("hostname").(string)
	N := 4
	for i := 0; i < N; i++ {
		requestURL := fmt.Sprintf("http://%s/successor", nextToAsk)

		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			log.Printf("error creating http get request: %s\n", err)
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			//log.Printf("error executing http get request: %s\n", err)
			//successor crashed -> find my new successor
			return
		}
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("error making http request: %s\n", err)
		}
		nextToAsk = string(resBody)
		if nextToAsk == (*ctx).Value("hostname").(string) {
			break
		}
		if string(resBody) == "" {
			f.successorStack.Push((*ctx).Value("hostname").(string))
		}
	}

}

func (f *FingerTable) StartEntryUpdater(ctx *context.Context) {
	//c := make(chan FingerTable)

	go func() {
		for {
			//running stabilizer
			f.stabilize(ctx)
			f.fixSuccessorList(ctx)
			f.fixFingers(ctx)
			f.checkPredecessorAlive(ctx)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func checkIfNodeIsAlive(url string) bool {
	requestURL := fmt.Sprintf("http://%s/node-info", url)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Printf("error creating http get request: %s\n", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error making http request: %s\n", err)
	}
	if string(resBody) == "" {
		return false
	}
	return true
}

//                =====================
//                ======= STACK =======
//                =====================
//FIXME maybe move to an own file later

type stack struct {
	elems []string
}

func (s *stack) Free() {
	s.elems = []string{}
}

func (s *stack) Push(e string) {
	s.elems = append(s.elems, e)
}

func (s *stack) Pop() (string, error) {
	if len(s.elems) == 0 {
		return "", errors.New("empty stack")
	}
	ret := s.elems[len(s.elems)-1]
	s.elems = s.elems[:len(s.elems)-1]
	return ret, nil
}
