package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	// Uncomment this block to pass the first stage
	"net"
	"os"
	"sync"
)

type SetValue struct {
	Value   string
	PX      int64
	SetTime time.Time
}

var redis_map map[string]SetValue = make(map[string]SetValue)

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
	// test
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
		conn.Write([]byte(ping(resp_array)))
	}

	if strings.ToLower(resp_array[0]) == "echo" {
		echo_result, err := echo(resp_array)
		if err != nil {
			return err
		}
		conn.Write([]byte(echo_result))
	}

	if strings.ToLower(resp_array[0]) == "set" {
		set_result, err := set(resp_array, redis_map)
		if err != nil {
			return err
		}
		conn.Write([]byte(set_result))
	}

	if strings.ToLower(resp_array[0]) == "get" {
		get_result, err := get(resp_array)
		if err != nil {
			return err
		}
		conn.Write([]byte(get_result))
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

func echo(data []string) (string, error) {
	if len(data) != 2 {
		return "", errors.New("Error ECHO not valid.")
	}

	bulk_string_length := "$" + strconv.Itoa(len(data[1]))
	// 改行文字列が末尾にも必要なため、空白文字列を入れている。
	echo_string := strings.Join([]string{bulk_string_length, data[1], ""}, "\r\n")

	return echo_string, nil
}

func set(data []string, redis_map map[string]SetValue) (string, error) {
	if len(data) < 3 {
		return "", errors.New("Error SET not valid")
	}

	var px int64
	if len(data) == 5 && strings.ToLower(data[3]) == "px" {
		px, _ = strconv.ParseInt(data[4], 10, 64)
	}

	redis_map[data[1]] = SetValue{Value: data[2], PX: px, SetTime: time.Now()}
	return "+OK\r\n", nil
}

func get(data []string) (string, error) {
	if len(data) != 2 {
		return "", errors.New("Error GET not valid")
	}

	get_data := redis_map[data[1]]

	time_diff := time.Now().Sub(get_data.SetTime)
	log.Printf("time diff: %v", time_diff.Milliseconds())
	log.Printf("PX: %v", get_data.PX)
	if time_diff.Milliseconds() > get_data.PX && get_data.PX != 0 {
		return "$-1\r\n", nil
	}

	bulk_string_length := "$" + strconv.Itoa(len(get_data.Value))
	// 改行文字列が末尾にも必要なため、空白文字列を入れている。
	get_string := strings.Join([]string{bulk_string_length, get_data.Value, ""}, "\r\n")
	return get_string, nil
}
