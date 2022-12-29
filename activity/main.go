package main

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
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

}
