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
	"github.com/yusufatalay/SocketProgramming/reservation/helper"
	"github.com/yusufatalay/SocketProgramming/reservation/models"
)

var ROOMSERVERPORT string
var ACTIVITYSERVERPORT string

func main() {

	// open the config file to write this serve's information
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("cannot found project's dotenv file: %v\n", err)
	}

	ROOMSERVERPORT = os.Getenv("ROOMSERVERPORT")
	ACTIVITYSERVERPORT = os.Getenv("ACTIVITYSERVERPORT")

	// check if room server is active using health endpoint using the socket connection
	// if not, exit the program
	RoomServerConn, err := net.Dial("tcp", "localhost:"+ROOMSERVERPORT)
	if err != nil {
		log.Fatalf("Cannor connect to Room Server %s", err.Error())
	}

	_, err = RoomServerConn.Write([]byte("GET /health HTTP/1.0"))
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
	RoomServerConn.Close()
	// check if activity server is active using health endpoint using the socket connection
	// if not, exit the program
	ActivityServerConn, err := net.Dial("tcp", "localhost:"+ACTIVITYSERVERPORT)
	if err != nil {
		log.Fatalf("Cannot connect Activity Server %s", err.Error())
	}
	_, err = ActivityServerConn.Write([]byte("GET /health HTTP/1.0"))
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
	ActivityServerConn.Close()

	if len(os.Args) != 2 {
		// nolint
		panic(errors.New("Should only provide a port number."))
	}

	// create a tcp socket that listens localhost:PORT
	ln, err := net.Listen("tcp", "localhost:"+os.Args[1])
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
	ActivityServerConn, err := net.Dial("tcp", "localhost:"+ACTIVITYSERVERPORT)
	if err != nil {
		log.Fatalf("Cannot Connect Activity Server %s", err.Error())
	}

	RoomServerConn, err := net.Dial("tcp", "localhost:"+ROOMSERVERPORT)
	if err != nil {
		log.Fatalf("Cannot Connect Room Server %s", err.Error())
	}

	reqParts := strings.Split(req, "\n")
	method := strings.TrimSpace(strings.Split(reqParts[0], " ")[0])
	url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])
	response := ""
	defer func() {
		_, err := (*conn).Write([]byte(response))
		if err != nil {
			log.Fatal(err)
		}
		ActivityServerConn.Close()
		RoomServerConn.Close()
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
		req := helper.CreateHTTPRequest(fmt.Sprintf("/check?name=%s", paramsMap["activity"]))
		_, err := ActivityServerConn.Write([]byte(req))
		if err != nil {
			log.Fatal(err)
		}
		buf := make([]byte, 1024)
		_, err = ActivityServerConn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				log.Fatal(err)
			}
		}
		if !strings.Contains(string(buf), "200 OK") {
			response = helper.CreateHTTPResponse("Reservation", 404, "Not Found",
				"Database error.", "Activity not found.")

			return
		}
		// check room server if the input is valid
		req = helper.CreateHTTPRequest(fmt.Sprintf("/reserve?name=%s&day=%s&hour=%s&duration=%s",
			paramsMap["room"], paramsMap["day"], paramsMap["hour"], paramsMap["duration"]))

		_, err = RoomServerConn.Write([]byte(req))
		if err != nil {
			if err.Error() != "EOF" {
				log.Fatal(err)
			}
		}
		buf = make([]byte, 1024)
		_, err = RoomServerConn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(string(buf), "400 Bad Request") {
			response = helper.CreateHTTPResponse("Reservation", 400, "Bad Request",
				"Database error.", "Invalid input.")

			return
		}

		if strings.Contains(string(buf), "404 Not Found") {
			response = helper.CreateHTTPResponse("Reservation", 404, "Not Found",
				"Database error.", "Room not found.")

			return
		}

		if strings.Contains(string(buf), "403 Forbidden") {
			response = helper.CreateHTTPResponse("Reservation", 403, "Forbidden",
				"Database error.", "Room not available.")

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

			response = helper.CreateHTTPResponse("Reservation", 200, "OK",
				"Reservation successful.", fmt.Sprintf("Reservation Details:\r\nReservation ID: %d\r\nRoom: %s\r\nActivity: %s\r\nDay: %d\r\nHour: %d\r\nDuration: %d\r\n\r\n",
					resID, paramsMap["room"], paramsMap["activity"], dayint, hourint, durationint))

			return
		}
	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //       // unmarshall the json body to a struct
		var body models.RoomReservation

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Reservation", 400, "Bad Request",
				"Parser error.", err.Error())

			return
		}
		// check activity server if the activity exists make a socket request
		req = helper.CreateHTTPRequest(fmt.Sprintf("/check?name=%s", body.ActivityName))
		_, err = ActivityServerConn.Write([]byte(req))
		if err != nil {
			log.Fatal(err)
		}
		buf := make([]byte, 1024)
		_, err = ActivityServerConn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if !strings.Contains(string(buf), "200 OK") {
			response = helper.CreateHTTPResponse("Reservation", 404, "Not Found",
				"Database error.", "Activity not found.")

			return
		}

		// check room server if the input is valid
		req = helper.CreateHTTPRequest(fmt.Sprintf("/check?name=%s&day=%d&hour=%d&duration=%d",
			body.RoomName, body.Day, body.Hour, body.Duration))
		_, err = RoomServerConn.Write([]byte(req))
		if err != nil {
			log.Fatal(err)
		}
		buf = make([]byte, 1024)
		_, err = RoomServerConn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(string(buf), "400 Bad Request") {
			response = helper.CreateHTTPResponse("Reservation", 400, "Bad Request",
				"Parser error.", "Invalid input.")
			return
		}
		if strings.Contains(string(buf), "404 Not Found") {
			response = helper.CreateHTTPResponse("Reservation", 404, "Not Found",
				"Database error.", "Room not found.")

			return
		}

		if strings.Contains(string(buf), "403 Forbidden") {
			response = helper.CreateHTTPResponse("Reservation", 403, "Forbidden",
				"Database error.", "Room is already reserved.")
			return
		}

		if strings.Contains(string(buf), "200 OK") {
			// successful now create local reservation
			resID, err := models.CreateRoomReservation(&models.RoomReservation{
				RoomName:     body.RoomName,
				ActivityName: body.ActivityName,
				Day:          body.Day,
				Hour:         body.Hour,
				Duration:     body.Duration,
			})

			if err != nil {
				log.Fatal(err)
			}
			response = helper.CreateHTTPResponse("Reservation", 200, "OK",
				"Reservation successful.", fmt.Sprintf("Reservation Details:\r\nReservation ID: %d\r\nRoom: %s\r\nActivity: %s\r\nDay: %d\r\nHour: %d\r\nDuration: %d\r\n\r\n",
					resID, body.RoomName, body.ActivityName, body.Day, body.Hour, body.Duration))

			return
		}
	}
}

func HandleListAvailability(conn *net.Conn, req string) {
	return
}

func HandleDisplay(conn *net.Conn, req string) {
	panic("unimplemented")
}
