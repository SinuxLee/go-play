package main

import (
	"encoding/json"
	"log"

	"play/internal/codec"
	"play/internal/schema"

	"github.com/google/gops/agent"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	if err := agent.Listen(agent.Options{
		ShutdownCleanup: true,
	}); err != nil {
		log.Fatal(err)
	}

	h := NewEventHandler(&codec.SimpleCodec{}, MakeMessageFun(func(s *Session, data []byte) error {
		p := &schema.Player{}
		if err := json.Unmarshal(data, p); err != nil {
			return err
		}

		return s.handler.Send(s.conn, data)
	}))
	h.Run("tcp://:8888")
}
