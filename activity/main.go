package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
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

	response := "HTTP/1.1 200 OK\n"
	response += "Content-Type: text/plain\n"
	response += "\n"
	response += "Activity Server is healty at port -> " + os.Getenv("ACTIVITYSERVERPORT")
	response += "\n"
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
		fmt.Println("GET request made to /add")
		url := strings.TrimSpace(strings.Split(reqParts[0], " ")[1])
		// GET request has been made, read the query params from the URL
		// url will be in the format of .../add?name=roomname so splitting according to the equal sign makes sense

		activityname := strings.Split(url, "=")[1]
		log.Println("activity" + activityname)
		// return error if roomname has not given
		if activityname == "" {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty /add?name=<activityname>"
			response += "\n"
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
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Activity already exists"
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
		response += fmt.Sprintf("Activity %s successfully added to database", activityname)
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
		var body models.Activity
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}
		fmt.Printf("activity object  %+v", body)
		err = models.CreateActivity(&body)
		if err != nil {
			if err.Error() == "Activity already exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Activity already exists"
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
		response += fmt.Sprintf("Activity %s successfully added to database", body.Name)
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

		activityname := strings.Split(url, "=")[1]
		log.Println("roomname " + activityname)
		// return error if roomname has not given
		if activityname == "" {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty /remove?name=<activity>"
			response += "\n"
			return
		}

		err := models.RemoveActivity(activityname)
		if err != nil {
			if err.Error() == "activity does not exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Activity does not exists"
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
		response += fmt.Sprintf("Activity %s successfully removed from database", activityname)
		response += "\n"
		return

	case "POST":
		// POST request has been made, read the request body
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct

		var body models.Activity
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}
		fmt.Printf("Activity object  %+v", body)
		err = models.RemoveActivity(body.Name)
		if err != nil {
			if err.Error() == "activity does not exists" {
				response += "HTTP/1.1 403 Forbidden\n"
				response += "Content-Type: text/plain\n"
				response += "\n"
				response += "Activity does not exists"
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
		response += fmt.Sprintf("Activity %s successfully removed from database", body.Name)
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
		log.Println("roomname " + activityname)
		// return error if roomname has not given
		if activityname == "" {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "name parameter is empty /check?name=<activity>"
			response += "\n"
			return
		}

		exists, err := models.CheckActivity(activityname)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Database Error:\t" + err.Error()
			response += "\n"
			return

		}

		if !exists {
			response += "HTTP/1.1 404 Not Found\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Activity does not exists"
			response += "\n"
			return
		}

		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += fmt.Sprintf("Activity %s exists in the database", activityname)
		response += "\n"
		return

	case "POST":
		// POST request has been made, read the request body
		reqBody := strings.Split(req, "{")[1]
		reqBody = "{" + reqBody
		fmt.Println(reqBody)
		fmt.Println(string([]byte(reqBody)))
		reqBodyByte := bytes.Trim([]byte(reqBody), "\x00") //		// unmarshall the json body to a struct

		var body models.Activity
		err := json.Unmarshal(reqBodyByte, &body)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Parser Error:\t" + err.Error()
			response += "\n"
			return
		}
		fmt.Printf("Activity object  %+v", body)

		exists, err := models.CheckActivity(body.Name)
		if err != nil {
			response += "HTTP/1.1 400 Bad Request\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Database Error:\t" + err.Error()
			response += "\n"
			return

		}

		if !exists {
			response += "HTTP/1.1 404 Not Found\n"
			response += "Content-Type: text/plain\n"
			response += "\n"
			response += "Activity does not exists"
			response += "\n"
			return
		}

		response += "HTTP/1.1 200 OK\n"
		response += "Content-Type: text/plain\n"
		response += "\n"
		response += fmt.Sprintf("Activity %s exists in the database", body.Name)
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
