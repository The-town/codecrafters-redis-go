package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
	"sync"
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
		wg := new(sync.WaitGroup)
		wg.Add(10)
		channel := make(chan error)

		for i := 0; i < 10; i++ {
			go run_recieve_process(l, wg, channel)
		}
		wg.Wait()
	}
}

func run_recieve_process(l net.Listener, wg *sync.WaitGroup, channel chan error) {

	conn := get_connection(l)
	for {
		log.Printf("conn %v", conn)
		err := recieve_data(conn, l)

		if err != nil {
			conn.Close()
			log.Println(err)
			break
		}
	}
	wg.Done()
}

func get_connection(l net.Listener) net.Conn {
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	return conn
}

func recieve_data(conn net.Conn, l net.Listener) error {

	data := make([]byte, 1024)
	count, err := conn.Read(data)
	if err != nil {
		return err
	}

	resp_array := get_resp_array(data[:count])
	if len(resp_array) == 0 {
		return errors.New("no array")
	}
	if resp_array[0] == "ping" {
		log.Printf("resp array %v", resp_array)
		conn.Write([]byte(ping(resp_array)))
	}
	return nil
}

func get_resp_array(data []byte) []string {
	// 配列を解析する関数
	split_data := strings.Split(string(data), "\r\n")

	// array_size, _ := strconv.Atoi(split_data[0][1:])
	resp_array := []string{}

	for _, d := range split_data[1:] {
		if strings.Index(d, "$") == 0 || d == "" {
			continue
		}
		resp_array = append(resp_array, d)
	}
	return resp_array
}

func ping(data []string) string {
	if len(data) == 1 {
		return "$4\r\nPONG\r\n"
	}
	return strings.Join(data[1:], " ")
}
