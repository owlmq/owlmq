package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/owlmq/owlmq/chord"
	"github.com/owlmq/owlmq/utils"
)

func New(ctx *context.Context, c chord.Chord_layer) Web_layer {
	return &WebServer{
		ctx:         ctx,
		chord_layer: c,
	}
}

type WebServer struct {
	mu          sync.Mutex
	ctx         *context.Context
	chord_layer chord.Chord_layer
}

func (ws *WebServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val, err := ws.chord_layer.Get(vars["key"])
	if err != nil {
		log.Println("Error", err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	fmt.Fprintf(w, "%v", val)
}

func (ws *WebServer) PutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	requestData, err := io.ReadAll(r.Body)

	err = ws.chord_layer.Put(vars["key"], string(requestData))
	if err != nil {
		log.Printf("Error %s\n", err)
	}
	fmt.Fprintf(w, "%v", string(requestData))
}

func (ws *WebServer) NodeInfoHandler(w http.ResponseWriter, r *http.Request) {
	type NodeInfo struct {
		Node_hash   string   `json:"node_hash"`
		Successor   string   `json:"successor"`
		Predecessor string   `json:"predecessor"`
		Others      []string `json:"others"`
	}

	nf := NodeInfo{
		Node_hash:   utils.GenerateSHA1((*ws.ctx).Value("hostname").(string)),
		Successor:   (*ws.ctx).Value("successor").(string),
		Predecessor: (*ws.ctx).Value("predecessor").(string),
		Others:      ws.chord_layer.ShowFingerTable(),
	}

	bs, err := json.MarshalIndent(nf, "", "    ")
	if err != nil {
		log.Printf(err.Error())
	}
	w.Write(bs)
}

// UpdatePredecessorHandler is needed for the stabilize
func (ws *WebServer) UpdatePredecessorHandler(w http.ResponseWriter, r *http.Request) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	body, _ := io.ReadAll(r.Body)

	if string(body) == "" {
		(*ws.ctx) = context.WithValue((*ws.ctx), "predecessor", (*ws.ctx).Value("hostname").(string))
		//fmt.Println("Predecessor was set to:", (*ws.ctx).Value("hostname").(string))
	} else {
		(*ws.ctx) = context.WithValue((*ws.ctx), "predecessor", string(body))
		//fmt.Println("Predecessor was set to:", string(body))
	}
	w.Write([]byte(""))
}

// PredecessorHandler is needed for the stabilize
func (ws *WebServer) PredecessorHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte((*ws.ctx).Value("predecessor").(string)))
}

// UpdateSuccessorHandler is needed for the leave
func (ws *WebServer) UpdateSuccessorHandler(w http.ResponseWriter, r *http.Request) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	body, _ := io.ReadAll(r.Body)

	if string(body) == "" {
		(*ws.ctx) = context.WithValue((*ws.ctx), "successor", (*ws.ctx).Value("hostname").(string))
		//fmt.Println("Successor was set to:", (*ws.ctx).Value("hostname").(string))
	} else {
		(*ws.ctx) = context.WithValue((*ws.ctx), "successor", string(body))
		//fmt.Println("Successor was set to:", string(body))
	}
	w.Write([]byte(""))
}

// SuccessorHandler is needed for the stabilize
func (ws *WebServer) SuccessorHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte((*ws.ctx).Value("successor").(string)))
}

type RequestNode struct {
	JoiningNode string `json:"joining_node"`
	Successor   string `json:"successor"`
}

func (ws *WebServer) FindSuccessorHandler(w http.ResponseWriter, r *http.Request) {
	var rn RequestNode

	//decode body if there is any
	if r.ContentLength != 0 {
		err := json.NewDecoder(r.Body).Decode(&rn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	jnkey := utils.GenerateSHA1(rn.JoiningNode)
	host := (*ws.ctx).Value("hostname").(string)
	suc := (*ws.ctx).Value("successor").(string)

	if suc == "" || utils.IsKeyInRange(jnkey, suc, host) {
		//fmt.Println("IN RANGE")
		rn.Successor = host
		tosendback, err := json.MarshalIndent(rn, "", "    ")
		if err != nil {
			log.Printf(err.Error())
		}
		w.Write(tosendback)
	} else {
		//fmt.Println("NOT IN RANGE")
		resBody := ws.chord_layer.FindSuccessor(rn.JoiningNode, suc)
		w.Write([]byte(resBody))
	}
}

func (ws *WebServer) JoinHandler(w http.ResponseWriter, r *http.Request) {
	//nprime -> known node inside the chord circle
	nprime := r.URL.Query().Get("nprime")

	ws.mu.Lock()
	defer ws.mu.Unlock()

	//TODO check if node already exists in ring

	//my successor is the next one to ask
	hostname := (*ws.ctx).Value("hostname").(string)

	//jsoc is the successor of the joining node
	jsuc := ws.chord_layer.FindSuccessor(hostname, nprime)

	//override the existing successor
	var rn RequestNode
	if err := json.Unmarshal([]byte(jsuc), &rn); err != nil {
		log.Printf("unable to Unmarshal result %s", err.Error())
	}
	(*ws.ctx) = context.WithValue((*ws.ctx), "successor", rn.Successor)

	w.Write([]byte(jsuc))
}

func (ws *WebServer) LeaveHandler(w http.ResponseWriter, r *http.Request) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	//fmt.Println("LEAVING")

	//TODO this is insecure since an attacker could simply update the successor and predecessor to join the cluster
	//notify successor that my predecessor is now his
	requestURL := fmt.Sprintf("http://%s/predecessor", (*ws.ctx).Value("successor").(string))
	req, err := http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer([]byte((*ws.ctx).Value("predecessor").(string))))
	if err != nil {
		log.Printf("error creating http get request: %s\n", err)
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("error executing http get request: %s\n", err)
		return
	}
	//notify predecessor that my successor is now his
	requestURL = fmt.Sprintf("http://%s/successor", (*ws.ctx).Value("predecessor").(string))
	req, err = http.NewRequest(http.MethodPut, requestURL, bytes.NewBuffer([]byte((*ws.ctx).Value("successor").(string))))
	if err != nil {
		log.Printf("error creating http get request: %s\n", err)
	}
	client = &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("error executing http get request: %s\n", err)
		return
	}

	//TODO we need to transfer the key-value pairs to the next node

	//clear successor stack, predecessor, successor, fingerTable
	cl := chord.New(ws.ctx)
	ws.chord_layer = cl
	(*ws.ctx) = context.WithValue((*ws.ctx), "predecessor", (*ws.ctx).Value("hostname").(string))
	(*ws.ctx) = context.WithValue((*ws.ctx), "successor", (*ws.ctx).Value("hostname").(string))

}

func (ws *WebServer) NewWebserver(ctx *context.Context) *mux.Router {
	r := mux.NewRouter()

	//data operations handlers
	r.HandleFunc("/storage/{key}", ws.GetHandler).Methods("GET")
	r.HandleFunc("/storage/{key}", ws.PutHandler).Methods("PUT")

	//network operations handlers
	r.HandleFunc("/join", ws.JoinHandler).Methods("POST")
	r.HandleFunc("/leave", ws.LeaveHandler).Methods("POST")

	//info handlers
	r.HandleFunc("/node-info", ws.NodeInfoHandler).Methods("GET")
	r.HandleFunc("/predecessor", ws.PredecessorHandler).Methods("GET")
	r.HandleFunc("/predecessor", ws.UpdatePredecessorHandler).Methods("PUT")
	r.HandleFunc("/successor", ws.SuccessorHandler).Methods("GET")
	r.HandleFunc("/successor", ws.UpdateSuccessorHandler).Methods("PUT")
	r.HandleFunc("/findsuccessor", ws.FindSuccessorHandler).Methods("POST")
	return r
}
