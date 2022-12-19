package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/yusufatalay/SocketProgramming/room/models"
)

// to handle concurrent request we need to create a mutex
var mutex sync.Mutex

func main() {

	// open the config file to write this serve's information
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("cannot found project's dotenv file: %v\n", err)
	}
	if len(os.Args) != 2 {
		panic(errors.New("Should only provide a port number."))
	}

	// insert this server's port number to config file
	godotenv.Write(map[string]string{
		"ROOMSERVERPORT": os.Args[1],
	}, "../.env")

	// register handler functions for each endpoint
	http.HandleFunc("/", health)
	http.HandleFunc("/add", HandleAdd)
	http.HandleFunc("/remove", HandleRemove)
	http.HandleFunc("/reserve", HandleReserve)
	http.HandleFunc("/checkavailablity", HandleCheckAvailability)
	// Create a http server with given port number that runs on localhost, this is a blocking method
	err = http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		panic(err)
	}

}

// health is for letting the controller server to know that this server is active
func health(w http.ResponseWriter, r *http.Request) {
	// just return 200 OK to let the caller know this server is active
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Room Server Is healty at port -> "+os.Getenv("ROOMSERVERPORT"))
	return
}

// HandleAdd will be triggerred when localhost/add has been visited
func HandleAdd(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// GET request has been made, read the query params from the URL
		io.WriteString(w, "GET request made to /add")
		roomname := r.URL.Query().Get("name")
		// return error if roomname has not given
		if roomname == "" {
			io.WriteString(w, "name parameter is empty /add?name=<roomname>")
		}
		io.WriteString(w, "roomname is /add "+roomname)
	case "POST":
		// POST request has been made, read the request body
		io.WriteString(w, "POST request made to /add")

		// unmarshall the json body to a struct
		defer r.Body.Close()
		body := struct {
			Name string `json:"name"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		io.WriteString(w, "in body /add is "+body.Name)

	default:
		// unknown http method has been used, throw error
		io.WriteString(w, "unkown http method used on /add")
	}

	return
}

func HandleRemove(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		// GET request has been made, read the query params from the URL
		io.WriteString(w, "GET request made to /remove")
		roomname := r.URL.Query().Get("name")
		// return error if roomname has not given
		if roomname == "" {
			io.WriteString(w, "name parameter is empty /remove?name=<roomname>")
		}
		io.WriteString(w, "roomname is /remove "+roomname)
	case "POST":
		// POST request has been made, read the request body
		io.WriteString(w, "POST request made to /remove")

		// unmarshall the json body to a struct
		defer r.Body.Close()
		body := struct {
			Name string `json:"name"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		io.WriteString(w, "in body /remove is "+body.Name)

	default:
		// unknown http method has been used, throw error
		io.WriteString(w, "unkown http method used on /remove")
	}

	return
}

func HandleReserve(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		// GET request has been made, read the query params from the URL
		io.WriteString(w, "GET request made to /reserve")
		roomname := r.URL.Query().Get("name")
		// return error if roomname has not given
		if roomname == "" {
			io.WriteString(w, "name parameter is empty /reserve?name=<roomname>&day=<day>&hour=<hour>&duration=<duration>")
		}
		day := r.URL.Query().Get("day")
		if day == "" {
			io.WriteString(w, "day parameter is empty /reserve?name=<roomname>&day=<day>&hour=<hour>&duration=<duration>")
		}
		hour := r.URL.Query().Get("hour")
		if hour == "" {
			io.WriteString(w, "hour parameter is empty /reserve?name=<roomname>&day=<day>&hour=<hour>&duration=<duration>")
		}
		duration := r.URL.Query().Get("duration")
		if duration == "" {
			io.WriteString(w, "duration parameter is empty /reserve?name=<roomname>&day=<day>&hour=<hour>&duration=<duration>")
		}

	case "POST":
		// POST request has been made, read the request body
		io.WriteString(w, "POST request made to /reserve")

		// unmarshall the json body to a struct
		defer r.Body.Close()
		body := struct {
			Name     string `json:"name"`
			Day      int    `json:"day"`
			Hour     int    `json:"hour"`
			Duration int    `json:"duration"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		io.WriteString(w, "in body /reserve is "+fmt.Sprintf("%+v", body))

	default:
		// unknown http method has been used, throw error
		io.WriteString(w, "unkown http method used on /reserve")
	}

	return
}

func HandleCheckAvailability(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		// GET request has been made, read the query params from the URL
		io.WriteString(w, "GET request made to /checkavailability")
		roomname := r.URL.Query().Get("name")
		// return error if roomname has not given
		if roomname == "" {
			io.WriteString(w, "name parameter is empty /checkavailability?name=<roomname>&day=<day>")
		}
		day := r.URL.Query().Get("day")
		// return error if day has not given
		if day == "" {
			io.WriteString(w, "name parameter is empty /checkavailability?name=<roomname>&day=<day>")
		}
		io.WriteString(w, "room and day is /checkavailability "+roomname+" "+day)
	case "POST":
		// POST request has been made, read the request body
		io.WriteString(w, "POST request made to /checkavailability")

		// unmarshall the json body to a struct
		defer r.Body.Close()
		body := struct {
			Name string `json:"name"`
			Day  int    `json:"day"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		io.WriteString(w, "in body /checkavailability is "+fmt.Sprintf("%+v", body))

	default:
		// unknown http method has been used, throw error
		io.WriteString(w, "unkown http method used on /checkavailability")
	}

	return
}
