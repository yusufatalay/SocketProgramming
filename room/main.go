package main

import (
	"errors"
	"net/http"
	"os"
	"sync"
)

// to handle concurrent request we need to create a mutex
var mutex sync.Mutex

func main() {

	if len(os.Args) != 2 {
		panic(errors.New("Should only provide a port number."))
	}

	// Create a http server with given port number that runs on localhost
	err := http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		panic(err)
	}

}

// this endpoint is for the parent server to check if this server is running

func health(w http.ResponseWriter, r *http.Request) {
	// just return 200 OK to let the caller know this server is active
	w.WriteHeader(http.StatusOK)
	return
}

func HandleAddGET(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleAddPOST(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleRemoveGET(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleRemovePOST(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleReserveGET(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleReservePOST(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleCheckAvailabilityGET(w http.ResponseWriter, r *http.Request) {
	return
}

func HandleCheckAvailabilityPOST(w http.ResponseWriter, r *http.Request) {
	return
}