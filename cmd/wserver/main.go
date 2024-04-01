package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/gorilla/websocket"
)

var (
	addr = ""
)

func init() {
	flag.StringVar(&addr, "addr", "localhost:8080", "http service address")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

type Handler struct {
	upgrader websocket.Upgrader
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	go h.process(c)
}

func (h *Handler) process(c *websocket.Conn) {
	defer c.Close()
	for {
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		t, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("recv: %s", message)
		if err := c.WriteMessage(t, message); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}

func main() {
	h := &Handler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  256,
			WriteBufferSize: 256,
			WriteBufferPool: &sync.Pool{},
		},
	}
	http.Handle("/ws", h)
	log.Fatal(http.ListenAndServe(addr, nil))
}
