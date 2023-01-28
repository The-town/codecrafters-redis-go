package main

import (
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		data := make([]byte, 1024)
		count, err := conn.Read(data)
		if err != nil {
			fmt.Println(err)
		}

		splited_data := strings.Split(string(data[:count]), " ")
		if splited_data[0] == "PING" {
			if len(splited_data) == 1 {
				conn.Write([]byte("PONG"))
			} else {
				conn.Write([]byte(strings.Join(splited_data[1:], " ")))
			}
		}
	}
}
