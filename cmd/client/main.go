package main

import (
	"log"
	"net"
	"time"
)

const numClients = 50000

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func sendRequest() {
	conn, err := net.Dial("tcp", "10.21.10.172:8888")
	if err != nil {
		log.Printf("Error connecting: %v\n", err.Error())
		return
	}
	defer conn.Close()

	buffer := make([]byte, 64)
	data := []byte("hello")
	for {
		_, err := conn.Write(data)
		if err != nil {
			log.Printf("Error writing to server: %v\n", err.Error())
			return
		}

		_, err = conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from server: %v\n", err.Error())
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {
	for i := 0; i < numClients; i++ {
		go sendRequest()
		time.Sleep(10 * time.Millisecond)
	}

	select {}
}
