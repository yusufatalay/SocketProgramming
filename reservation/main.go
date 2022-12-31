package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/yusufatalay/SocketProgramming/reservation/models"
)

var RoomServerConn net.Conn
var ActivityServerConn net.Conn

func main() {

	// open the config file to write this serve's information
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("cannot found project's dotenv file: %v\n", err)
	}

	var ROOMSERVERPORT = os.Getenv("ROOMSERVERPORT")
	var ACTIVITYSERVERPORT = os.Getenv("ACTIVITYSERVERPORT")

	// check if room server is active using health endpoint using the socket connection
	// if not, exit the program
	RoomServerConn, err = net.Dial("tcp", "localhost:"+ROOMSERVERPORT)
	if err != nil {
		log.Fatal(err)
	}
	defer RoomServerConn.Close()
	_, err = RoomServerConn.Write([]byte("GET /health HTTP/1.1"))
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 1024)
	_, err = RoomServerConn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(string(buf), "200 OK") {
		fmt.Println("Room server is not active.")

		return
	}
	// check if activity server is active using health endpoint using the socket connection
	// if not, exit the program
	ActivityServerConn, err = net.Dial("tcp", "localhost:"+ACTIVITYSERVERPORT)
	if err != nil {
		log.Fatal(err)
	}
	defer ActivityServerConn.Close()
	_, err = ActivityServerConn.Write([]byte("GET /health HTTP/1.1"))
	if err != nil {
		log.Fatal(err)
	}

	buf = make([]byte, 1024)
	_, err = ActivityServerConn.Read(buf)
	if err != nil {

		log.Fatal(err)
	}
	if !strings.Contains(string(buf), "200 OK") {
		fmt.Println("Activity server is not active.")

		return
	}

	if len(os.Args) != 2 {
		// nolint
		panic(errors.New("Should only provide a port number."))
	}

	// insert this server's port number to config file
	err = godotenv.Write(map[string]string{
		"RESERVATIONSERVERPORT": os.Args[1],
	}, "../.env")
	if err != nil {
		log.Fatalf("cannot write to project's dotenv file: %v\n", err)
	}

	// create a tcp socket that listens localhost:PORT
	ln, err := net.Listen("tcp", "localhost:"+os.Getenv("RESERVATIONSERVERPORT"))
	if err != nil {
		log.Fatal(err)
	}
	// deferred close the connection
	defer ln.Close()

	//	program loop
	for {
		// accept connection
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// create new thread for each connection request to handle concurrency
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	// Use bufio to read the request
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)

		return
	}

	// Parse the request

	reqStr := string(buf)
	reqParts := strings.Split(reqStr, "\n")
	url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])

	if strings.HasPrefix(url, "/reserve") {
		HandleReserve(&conn, reqStr)

	} else if strings.HasPrefix(url, "/listavailibility") {
		HandleListAvailability(&conn, reqStr)

	} else if strings.HasPrefix(url, "/display") {
		HandleDisplay(&conn, reqStr)

	}
}

func HandleReserve(conn *net.Conn, req string) {
	reqParts := strings.Split(req, "\n")
	method := strings.TrimSpace(strings.Split(reqParts[0], " ")[0])
	url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])
	response := ""
	defer func() {
		_, err := (*conn).Write([]byte(response))
		if err != nil {
			log.Fatal(err)
		}
		(*conn).Close()
	}()
	switch method {
	case "GET":
		// GET request has been made, read the query params from the URL
		query := strings.Split(url, "?")[1]
		params := strings.Split(query, "&")
		paramsMap := make(map[string]string)
		for _, param := range params {
			temp := strings.Split(param, "=")
			paramsMap[temp[0]] = temp[1]
		}
		// check activity server if the activity exists make a socket request
		_, err := ActivityServerConn.Write([]byte(fmt.Sprintf("GET /check?name=" + paramsMap["activity"] + " HTTP/1.1")))
		if err != nil {
			log.Fatal(err)
		}
		buf := make([]byte, 1024)
		_, err = ActivityServerConn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if !strings.Contains(string(buf), "200 OK") {
			response += "HTTP/1.1 404 Not Found\n"
			response += "Content-Type: text/plain\n"
			response += "Activity not found.\n"

			return
		}
		// check room server if the input is valid
		_, err = RoomServerConn.Write([]byte(fmt.Sprintf("GET /reserve?name=%s&day=%s&hour=%s&duration=%s HTTP/1.1",
			paramsMap["room"], paramsMap["day"], paramsMap["hour"], paramsMap["duration"])))
		if err != nil {
			log.Fatal(err)
		}
		buf = make([]byte, 1024)
		_, err = RoomServerConn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(string(buf), "400 Bad Request") {
			response = string(buf)
			return
		}

		if strings.Contains(string(buf), "404 Not Found") {
			response += "HTTP/1.1 404 Not Found\n"
			response += "Content-Type: text/plain\n"
			response += "Room not found.\n"

			return
		}

		if strings.Contains(string(buf), "403 Forbidden") {
			response = string(buf)
			return
		}
		if strings.Contains(string(buf), "200 OK") {
			// successful now create local reservation
			dayint, _ := strconv.Atoi(paramsMap["day"])
			hourint, _ := strconv.Atoi(paramsMap["hour"])
			durationint, _ := strconv.Atoi(paramsMap["duration"])
			resID, err := models.CreateRoomReservation(&models.RoomReservation{
				RoomName:     paramsMap["room"],
				ActivityName: paramsMap["activity"],
				Day:          dayint,
				Hour:         hourint,
				Duration:     durationint,
			})

			if err != nil {
				log.Fatal(err)
			}

			response += "HTTP/1.1 200 OK\n"
			response += "Content-Type: text/plain\n"
			response += "Reservation successful.\n"
			response += "Reservation Details:\n"
			response += fmt.Sprintf("Reservation ID: %d\n", resID)
			response += "Room: " + paramsMap["room"] + "\n"
			response += "Activity: " + paramsMap["activity"] + "\n"
			response += "Day: " + paramsMap["day"] + "\n"
			response += "Hour: " + paramsMap["hour"] + "\n"
			response += "Duration: " + paramsMap["duration"] + "\n"

			return
		}
	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct
		var body models.RoomReservation

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"

			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"

			return
		}
	}
}

func HandleListAvailability(conn *net.Conn, req string) {
	panic("unimplemented")
}

func HandleDisplay(conn *net.Conn, req string) {
	panic("unimplemented")
}
