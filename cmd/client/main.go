package main

import (
	"flag"
	"log"
	"time"

	"github.com/panjf2000/gnet/v2"
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
	handler := NewEventHandler()
	cli, err := gnet.NewClient(
		handler,
		gnet.WithTicker(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithLockOSThread(true),
		gnet.WithMulticore(true),
	)

	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	cli.Start()

	for i := 0; i < numClients; i++ {
		_, err := cli.Dial("tcp", addr)
		if err != nil {
			log.Printf("Error connecting: %v\n", err.Error())
			continue
		}

		time.Sleep(time.Millisecond)
	}

	select {}
}
