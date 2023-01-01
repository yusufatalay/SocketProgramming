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
	"github.com/yusufatalay/SocketProgramming/activity/helper"
	"github.com/yusufatalay/SocketProgramming/activity/models"
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
		"ROOMSERVERPORT":     os.Getenv("ROOMSERVERPORT"),
		"ACTIVITYSERVERPORT": os.Args[1],
	}, "../.env")

	// create a tcp socket that listens localhost:PORT
	ln, err := net.Listen("tcp", "localhost:"+os.Getenv("ACTIVITYSERVERPORT"))
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

	if strings.HasPrefix(url, "/health") {
		health(&conn)
	} else if strings.HasPrefix(url, "/add") {
		HandleAdd(&conn, reqStr)
	} else if strings.HasPrefix(url, "/remove") {
		HandleRemove(&conn, reqStr)
	} else if strings.HasPrefix(url, "/check") {
		HandleCheck(&conn, reqStr)
	}

}

func health(conn *net.Conn) {
	// just return 200 OK to let the caller know this server is active

	response := "HTTP/1.0 200 OK\r\n"
	response += "Content-Type: text/plain\r\n"
	response += "Content-Length: " + strconv.Itoa(len([]byte(fmt.Sprintf("Activity Server is healty at port -> "+os.Getenv("ACTIVITYSERVERPORT")))))
	response += "\r\n\r\n"
	response += "Activity Server is healty at port -> " + os.Getenv("ACTIVITYSERVERPORT")
	response += "\r\n"
	_, err := (*conn).Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}

	(*conn).Close()

}

// HandleAdd will be triggerred when localhost/add has been visited
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

		activityname := strings.Split(url, "=")[1]
		// return error if roomname has not given
		if activityname == "" {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Validation Error", "name parameter is empty /add?name=<activityname>")

			return
		}
		// create a room instance with the given name
		a := &models.Activity{
			Name: activityname,
		}
		// create the room in database
		err := models.CreateActivity(a)

		if err != nil {
			if err.Error() == "Activity already exists" {
				response = helper.CreateHTTPResponse("Activity", 403, "Forbidden",
					"Validation Error", "Activity already exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Activity", 200, "OK",
			"Succesfull", fmt.Sprintf("Activity %s successfully added to database", activityname))

		return

	case "POST":
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00")

		var body models.Activity
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}
		err = models.CreateActivity(&body)
		if err != nil {
			if err.Error() == "Activity already exists" {
				response = helper.CreateHTTPResponse("Activity", 403, "Forbidden",
					"Database Error", "Activity already exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Activity", 200, "OK",
			"Succesfull", fmt.Sprintf("Activity %s successfully added to database", body.Name))

		return
	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
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

		activityname := strings.Split(url, "=")[1]
		// return error if roomname has not given
		if activityname == "" {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Validation Error", "name parameter is empty /remove?name=<activityname>")

			return
		}

		err := models.RemoveActivity(activityname)
		if err != nil {
			if err.Error() == "activity does not exists" {
				response = helper.CreateHTTPResponse("Activity", 403, "Forbidden",
					"Database Error", "Activity does not exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Activity", 200, "OK",
			"Succesfull", fmt.Sprintf("Activity %s successfully removed from database", activityname))

		return

	case "POST":
		// POST request has been made, read the request body
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct

		var body models.Activity
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}
		err = models.RemoveActivity(body.Name)
		if err != nil {
			if err.Error() == "activity does not exists" {
				response = helper.CreateHTTPResponse("Activity", 403, "Forbidden",
					"Database Error", "Activity does not exists")

				return
			} else {
				response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
					"Database Error", err.Error())

				return
			}

		}
		response = helper.CreateHTTPResponse("Activity", 200, "OK",
			"Succesfull", fmt.Sprintf("Activity %s successfully removed from database", body.Name))

		return
	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
			"Method not supported", "Method not supported")

		return
	}
}

func HandleCheck(conn *net.Conn, req string) {

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

		activityname := strings.Split(url, "=")[1]
		// return error if roomname has not given
		if activityname == "" {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Validation Error", "name parameter is empty /check?name=<activityname>")

			return
		}

		exists, err := models.CheckActivity(activityname)
		if err != nil {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Database Error", err.Error())

			return

		}

		if !exists {
			response = helper.CreateHTTPResponse("Activity", 404, "Not Found",
				"Database Error", "Activity does not exists")

			return
		}

		response = helper.CreateHTTPResponse("Activity", 200, "OK",
			"Succesfull", fmt.Sprintf("Activity %s exists in the database", activityname))

		return

	case "POST":
		// POST request has been made, read the request body
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct

		var body models.Activity
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Parser Error", err.Error())

			return
		}

		exists, err := models.CheckActivity(body.Name)
		if err != nil {
			response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
				"Database Error", err.Error())

			return

		}

		if !exists {
			response = helper.CreateHTTPResponse("Activity", 404, "Not Found",
				"Database Error", "Activity does not exists")

			return
		}

		response = helper.CreateHTTPResponse("Activity", 200, "OK",
			"Succesfull", fmt.Sprintf("Activity %s exists in the database", body.Name))

		return
	default:
		// unknown http method has been used, throw error
		response = helper.CreateHTTPResponse("Activity", 400, "Bad Request",
			"Method not supported", "Method not supported")

		return
	}

}
