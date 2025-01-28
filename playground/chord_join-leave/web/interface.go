package web

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Web_layer interface {
	//old API routes
	GetHandler(w http.ResponseWriter, r *http.Request)
	PutHandler(w http.ResponseWriter, r *http.Request)
	NetworkHandler(w http.ResponseWriter, r *http.Request)

	//new API routes
	NodeInfoHandler(w http.ResponseWriter, r *http.Request)
	JoinHandler(w http.ResponseWriter, r *http.Request)
	LeaveHandler(w http.ResponseWriter, r *http.Request)

	NewWebserver(*context.Context) *mux.Router
}
