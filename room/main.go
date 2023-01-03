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
	"github.com/yusufatalay/SocketProgramming/room/helper"
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
		"ROOMSERVERPORT":     os.Args[1],
		"ACTIVITYSERVERPORT": os.Getenv("ACTIVITYSERVERPORT"),
	}, "../.env")

	// create a tcp socket that listens localhost:PORT
	ln, err := net.Listen("tcp", "localhost:"+os.Getenv("ROOMSERVERPORT"))
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
	} else if strings.HasPrefix(url, "/checkweeklyavailability") {
		HandleCheckWeeklyAvailability(&conn, reqStr)
	}

}

// health is for letting the controller server to know that this server is active.
func health(conn *net.Conn) {
	// just return 200 OK to let the caller know this server is active

	response := "HTTP/1.0 200 OK\r\n"
	response += "Content-Type: text/plain\r\n"
	response += "\r\n"
	response += "Room Server is healty at port -> " + os.Getenv("ROOMSERVERPORT")
	response += "\r\n"
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
		url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])
		// GET request has been made, read the query params from the URL
		// url will be in the format of .../add?name=roomname so splitting according to the equal sign makes sense

		roomname := strings.Split(url, "=")[1]
		// return error if roomname has not given
		if roomname == "" {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Empty Parameter", "name parameter is empty")

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
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room already exists", "There is already a room exists with the same name")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Room", 200, "OK",
			"Succesfull", fmt.Sprintf("Room %s successfully added to database", roomname))

		return

	case "POST":
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00")
		// POST request has been made, read the request body
		// unmarshall the json body to a struct
		var body models.Room
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}
		err = models.CreateRoom(&body)
		if err != nil {
			if err.Error() == "Room already exists" {
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room already exists", "There is already a room exists with the same name")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Room", 200, "OK",
			"Succesfull", fmt.Sprintf("Room %s successfully added to database", body.Name))

		return
	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
			"Method not supported", "Method not supported")

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
		// return error if roomname has not given
		if roomname == "" {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Empty Parameter", "name parameter is empty")

			return
		}

		err := models.RemoveRoom(roomname)
		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room does not exists", "There is no room exists with the given name")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Room", 200, "OK",
			"Succesfull", fmt.Sprintf("Room %s successfully removed from database", roomname))

		return

	case "POST":
		// POST request has been made, read the request body
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct

		var body models.Room
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}
		err = models.RemoveRoom(body.Name)
		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room does not exists", "There is no room exists with the given name")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Room", 200, "OK",
			"Succesfull", fmt.Sprintf("Room %s successfully removed from database", body.Name))

		return
	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
			"Method not supported", "Method not supported")

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
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Empty Parameter", "name parameter is empty")

			return
		} else if _, ok := paramsMap["day"]; !ok {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Empty Parameter", "day parameter is empty")

			return
		} else if _, ok := paramsMap["hour"]; !ok {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Empty Parameter", "hour parameter is empty")
			return
		} else if _, ok := paramsMap["duration"]; !ok {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Empty Parameter", "duration parameter is empty")

			return
		}

		dayint, err := strconv.Atoi(paramsMap["day"])
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "day value should be a number")

			return
		}
		hourint, err := strconv.Atoi(paramsMap["hour"])
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "hour value should be a number")

			return
		}

		durationint, err := strconv.Atoi(paramsMap["duration"])
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "duration value should be a number")

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
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room does not exists", "There is no room exists with the given name")

				return
			} else if err.Error() == "Already Reserved" {
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room already reserved", "Room already reserved in given time slice")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}
		}

		response = helper.CreateHTTPResponse("Room", 200, "OK",
			"Succesfull", "Reservation created successfully")

		return

	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct
		var body models.Reservation

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}

		err = models.CreateReservation(&body)
		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Room does not exists", "There is no room exists with the given name")

				return
			} else if err.Error() == "Already Reserved" {
				response = helper.CreateHTTPResponse("Room", 403, "Forbidden",
					"Room already reserved", "Room already reserved in given time slice")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}
		}

		response = helper.CreateHTTPResponse("Room", 200, "OK",
			"Succesfull", "Reservation created successfully")

		return

	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
			"Method not supported", "Method not supported")

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
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "name parameter is empty")

			return
		} else if _, ok := paramsMap["day"]; !ok {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "day parameter is empty")

			return
		}
		dayint, err := strconv.Atoi(paramsMap["day"])
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "day value should be a number")

			return
		}
		hours, err := models.GetAvailableHours(paramsMap["name"], dayint)

		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 404, "Not Found",
					"Room does not exists", "Room does not exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Validation Error", err.Error())

				return
			}
		}

		hoursStr := strings.Builder{}
		for _, h := range hours {
			hoursStr.WriteString(h + " ")
		}
		hoursStr.WriteRune('\n')

		response = helper.CreateHTTPResponse("Room", 200, "OK",
			fmt.Sprintf("Available hours for room %s for day %s is listed below",
				paramsMap["name"], paramsMap["day"]), hoursStr.String())

		return

	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct
		var body struct {
			Name string `json:"room_name"`
			Day  int    `json:"day"`
		}

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}

		hours, err := models.GetAvailableHours(body.Name, body.Day)

		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 404, "Not Found",
					"Room does not exists", "Room does not exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Validation Error", err.Error())

				return
			}
		}

		hoursStr := strings.Builder{}
		for _, h := range hours {
			hoursStr.WriteString(h + " ")
		}
		hoursStr.WriteRune('\n')

		response = helper.CreateHTTPResponse("Room", 200, "OK",
			fmt.Sprintf("Available hours for room %s for day %d is listed below",
				body.Name, body.Day), hoursStr.String())

		return

	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
			"Method not supported", "Method not supported")

		return
	}
}

func HandleCheckWeeklyAvailability(conn *net.Conn, req string) {
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
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", "name parameter is empty")

			return
		}

		daysandhours, err := models.GetAllAvailableHours(paramsMap["name"])

		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 404, "Not Found",
					"Room does not exists", "Room does not exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Validation Error", err.Error())

				return
			}
		}

		hoursStr := strings.Builder{}
		for day, hours := range daysandhours {
			hoursStr.WriteString(day + ": ")
			for _, h := range hours {
				hoursStr.WriteString(h + " ")
			}
			hoursStr.WriteRune('\n')
		}
		hoursStr.WriteRune('\n')

		response = helper.CreateHTTPResponse("Room", 200, "OK",
			fmt.Sprintf("Available hours for room %s for this week is listed below",
				paramsMap["name"]), hoursStr.String())

		return

	case "POST":
		// POST request has been made, read the request body

		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct
		var body struct {
			Name string `json:"room_name"`
		}

		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}

		daysandhours, err := models.GetAllAvailableHours(body.Name)

		if err != nil {
			if err.Error() == "Room does not exists" {
				response = helper.CreateHTTPResponse("Room", 404, "Not Found",
					"Room does not exists", "Room does not exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
					"Validation Error", err.Error())

				return
			}
		}

		hoursStr := strings.Builder{}
		for day, hours := range daysandhours {
			hoursStr.WriteString(day + ": ")
			for _, h := range hours {
				hoursStr.WriteString(h + " ")
			}
			hoursStr.WriteRune('\n')
		}
		hoursStr.WriteRune('\n')
		response = helper.CreateHTTPResponse("Room", 200, "OK",
			fmt.Sprintf("Available hours for room %s for this week is listed below",
				body.Name), hoursStr.String())

		return

	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Room", 400, "Bad Request",
			"Method not supported", "Method not supported")

		return
	}
}
