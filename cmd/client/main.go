package main

import (
	"flag"
	"log"
	"time"

	"github.com/google/gops/agent"
)

const (
	numClients = 10000
	bufferSize = 1024
)

var (
	addr = "10.21.10.172:8888"
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:8888", "--addr 127.0.0.1:8888")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true,
	}); err != nil {
		log.Fatal(err)
	}

	h, err := NewEventHandler()
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	for i := 0; i < numClients; i++ {
		_, err := h.Connect(addr)
		if err != nil {
			log.Printf("Error connecting: %v\n", err.Error())
			continue
		}

		time.Sleep(time.Millisecond)
	}

	select {}
}
