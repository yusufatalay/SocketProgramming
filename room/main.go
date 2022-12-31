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
	"github.com/yusufatalay/SocketProgramming/room/models"
)

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

	// create a tcp socket that listens localhost:PORT
	ln, err := net.Listen("tcp", "127.0.0.1:"+os.Getenv("ROOMSERVERPORT"))
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

		return
	}

	// Parse the request

	reqStr := string(buf)
	reqParts := strings.Split(reqStr, "\n")
	url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])

	// use appropriate handler for the url
	if strings.HasPrefix(url, "/health") {
		health(&conn)
	} else if strings.HasPrefix(url, "/add") {
		HandleAdd(&conn, reqStr)

	} else if strings.HasPrefix(url, "/remove") {
		HandleRemove(&conn, reqStr)

	} else if strings.HasPrefix(url, "/reserve") {
		HandleReserve(&conn, reqStr)

	} else if strings.HasPrefix(url, "/checkavailability") {
		HandleCheckAvailability(&conn, reqStr)
	}

}

// health is for letting the controller server to know that this server is active.
func health(conn *net.Conn) {
	// just return 200 OK to let the caller know this server is active

	response := "HTTP/1.1 200 OK\n"
	response += "Content-Type: text/plain\n"
	response += "\n"
	response += "Room Server is healty at port -> " + os.Getenv("ROOMSERVERPORT")
	response += "\n"
	_, err := (*conn).Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}

	(*conn).Close()
}

// HandleAdd will be triggerred when localhost/add has been visited.
func HandleAdd(conn *net.Conn, req string) {

	reqParts := strings.Split(req, "\n")
	method := strings.TrimSpace(strings.Split(reqParts[0], " ")[0])
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
		fmt.Println("GET request made to /add")
		url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])
		// GET request has been made, read the query params from the URL
		// url will be in the format of .../add?name=roomname so splitting according to the equal sign makes sense

		roomname := strings.Split(url, "=")[1]
		log.Println("roomname " + roomname)
		// return error if roomname has not given
		if roomname == "" {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty /add?name=<roomname>"
			response += "\n"
			return
		}
		// create a room instance with the given name
		r := &models.Room{
			Name: roomname,
		}
		// create the room in database
		err := models.CreateRoom(r)

		if err != nil {
			if err.Error() == "Room already exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room already exists"
				response += "\n"
				return
			} else {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Database Error:\t" + err.Error()
				response += "\n"
				return
			}

		}
		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += fmt.Sprintf("Room %s successfully added to database", roomname)
		response += "\n"
		return

	case "POST":
		fmt.Println("POST request made to /add")
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00")
		// POST request has been made, read the request body
		// unmarshall the json body to a struct
		var body models.Room
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}
		fmt.Printf("room object  %+v", body)
		err = models.CreateRoom(&body)
		if err != nil {
			if err.Error() == "Room already exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room already exists"
				response += "\n"
				return
			} else {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Database Error:\t" + err.Error()
				response += "\n"
				return
			}

		}
		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += fmt.Sprintf("Room %s successfully added to database", body.Name)
		response += "\n"
		return
	default:
		// unknown http method has been used, throw error
		response += "HTTP/1.1 400 Bad Request\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Method not supported"
		response += "\n"
		return
	}

}

func HandleRemove(conn *net.Conn, req string) {

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

		roomname := strings.Split(url, "=")[1]
		log.Println("roomname " + roomname)
		// return error if roomname has not given
		if roomname == "" {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty /remove?name=<roomname>"
			response += "\n"
			return
		}

		err := models.RemoveRoom(roomname)
		if err != nil {
			if err.Error() == "Room does not exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room does not exists"
				response += "\n"
				return
			} else {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Database Error:\t" + err.Error()
				response += "\n"
				return
			}

		}
		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += fmt.Sprintf("Room %s successfully removed from database", roomname)
		response += "\n"
		return

	case "POST":
		// POST request has been made, read the request body
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct

		var body models.Room
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}
		fmt.Printf("room object  %+v", body)
		err = models.RemoveRoom(body.Name)
		if err != nil {
			if err.Error() == "Room does not exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room does not exists"
				response += "\n"
				return
			} else {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Database Error:\t" + err.Error()
				response += "\n"
				return
			}

		}
		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += fmt.Sprintf("Room %s successfully removed from database", body.Name)
		response += "\n"
		return
	default:
		// unknown http method has been used, throw error
		response += "HTTP/1.1 400 Bad Request\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Method not supported"
		response += "\n"
		return
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
		// return error if roomname has not given
		if _, ok := paramsMap["name"]; !ok {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty"
			response += "\n"
			return
		} else if _, ok := paramsMap["day"]; !ok {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "day parameter is empty"
			response += "\n"
			return
		} else if _, ok := paramsMap["hour"]; !ok {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "hour parameter is empty"
			response += "\n"
			return
		} else if _, ok := paramsMap["duration"]; !ok {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "duration parameter is empty"
			response += "\n"
			return
		}

		dayint, err := strconv.Atoi(paramsMap["day"])
		if err != nil {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "day value should be a number"
			response += "\n"
			return
		}
		hourint, err := strconv.Atoi(paramsMap["hour"])

		if err != nil {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "hour value should be a number"
			response += "\n"
			return
		}
		durationint, err := strconv.Atoi(paramsMap["duration"])

		if err != nil {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "duration value should be a number"
			response += "\n"
			return
		}
		err = models.CreateReservation(&models.Reservation{
			RoomName: paramsMap["name"],
			Day:      dayint,
			Hour:     hourint,
			Duration: durationint,
		})
		if err != nil {
			if err.Error() == "Room does not exists" {
				response += "HTTP/1.1 404 Not Found\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room does not exists"
				response += "\n"
				return
			} else if err.Error() == "Already Reserved" {

				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room already reserved in given time slice"
				response += "\n"
				return
			} else {

				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Database error: " + err.Error()
				response += "\n"
				return
			}
		}

		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Reservation created successfully"
		response += "\n"
		return

	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct
		var body models.Reservation

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}

		err = models.CreateReservation(&body)
		if err != nil {
			if err.Error() == "Room does not exists" {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room does not exists"
				response += "\n"
				return
			} else if err.Error() == "Already Reserved" {

				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room already reserved in given time slice"
				response += "\n"
				return
			} else {

				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Database error: " + err.Error()
				response += "\n"
				return
			}
		}

		response += "HTTP/1.1 200 Bad Request\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Reservation created successfully"
		response += "\n"
		return

	default:
		// unknown http method has been used, throw error
		response += "HTTP/1.1 400 Bad Request\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Method not supported"
		response += "\n"
		return
	}

}

func HandleCheckAvailability(conn *net.Conn, req string) {
	reqParts := strings.Split(req, "\n")
	method := strings.TrimSpace(strings.Split(reqParts[0], " ")[0])
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
		url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])
		// GET request has been made, read the query params from the URL
		// url will be in the format of .../add?name=roomname so splitting according to the equal sign makes sense

		// GET request has been made, read the query params from the URL
		query := strings.Split(url, "?")[1]
		params := strings.Split(query, "&")
		paramsMap := make(map[string]string)
		for _, param := range params {
			temp := strings.Split(param, "=")
			paramsMap[temp[0]] = temp[1]
		}
		// return error if roomname has not given
		if _, ok := paramsMap["name"]; !ok {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty"
			response += "\n"
			return
		} else if _, ok := paramsMap["day"]; !ok {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "day parameter is empty"
			response += "\n"
			return
		}
		dayint, err := strconv.Atoi(paramsMap["day"])
		if err != nil {

			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "day value should be a number"
			response += "\n"
			return
		}
		hours, err := models.GetAvailableHours(paramsMap["name"], dayint)

		if err != nil {
			if err.Error() == "Room does not exists" {
				response += "HTTP/1.1 404 Not Found\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room does not exists"
				response += "\n"

				return
			} else {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Validation error: " + err.Error()
				response += "\n"

				return
			}
		}

		hoursStr := strings.Builder{}
		for _, h := range hours {
			hoursStr.WriteString(h + " ")
		}
		hoursStr.WriteRune('\n')

		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Available hours for room " + paramsMap["name"] + " for day " +
			paramsMap["day"] + " is listed below"
		response += "\n"
		response += hoursStr.String()
		response += "\n"
		return

	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct
		var body struct {
			Name string `json:"room_name"`
			Day  int    `json:"day"`
		}

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}

		hours, err := models.GetAvailableHours(body.Name, body.Day)

		if err != nil {
			if err.Error() == "Room does not exists" {
				response += "HTTP/1.1 404 Not Found\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Room does not exists"
				response += "\n"

				return
			} else {
				response += "HTTP/1.1 400 Bad Request\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Validation error: " + err.Error()
				response += "\n"

				return
			}
		}

		hoursStr := strings.Builder{}
		for _, h := range hours {
			hoursStr.WriteString(h + " ")
		}
		hoursStr.WriteRune('\n')

		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Available hours for room " + body.Name + " for day " +
			strconv.Itoa(body.Day) + " is listed below"
		response += "\n"
		response += hoursStr.String()
		response += "\n"
		return

	default:
		// unknown http method has been used, throw error
		response += "HTTP/1.1 400 Bad Request\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += "Method not supported"
		response += "\n"
		return
	}
}
