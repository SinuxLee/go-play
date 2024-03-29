package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"play/internal/codec"

	"github.com/google/gops/agent"
)

var (
	addr   = "10.21.10.172:8888"
	client = 0
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:8888", "--addr 127.0.0.1:8888")
	flag.IntVar(&client, "client", 1000, "--client 1000")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true,
	}); err != nil {
		log.Fatal(err)
	}

	h, err := NewEventHandler(&codec.SimpleCodec{}, MakeMessageFun(func(s *Session, data []byte) error {
		m := make(map[string]any)
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		// log.Printf("%+v\n", len(data))
		return nil
	}))

	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	for i := 0; i < client; i++ {
		_, err := h.Connect(addr)
		if err != nil {
			log.Printf("Error connecting: %v\n", err.Error())
			continue
		}

		time.Sleep(time.Millisecond)
	}

	select {}
}
