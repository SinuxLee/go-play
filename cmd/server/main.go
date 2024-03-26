package main

import (
	"log"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	h := NewEventHandler()
	h.Run("tcp://:8888")
}
