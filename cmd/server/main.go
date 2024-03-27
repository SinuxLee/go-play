package main

import (
	"log"

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

	h := NewEventHandler()
	h.Run("tcp://:8888")
}
