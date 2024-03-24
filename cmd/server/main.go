package main

import (
	"log"
	"net"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func handleConnection(conn net.Conn, idx int) {
	defer conn.Close()

	buffer := make([]byte, 64)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("%v reading: %v\n", idx, err)
			return
		}

		data := buffer[:n]
		_, err = conn.Write(data)
		if err != nil {
			log.Printf("%v writing: %v\n", idx, err)
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Printf("Error listening: %v\n", err)
		return
	}

	defer listener.Close()
	log.Println("Server started. Listening on :8888")

	idx := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		idx++

		log.Println("Accept client ", idx)
		go handleConnection(conn, idx)
	}
}
