package main

import (
	"flag"
	"log"
	"net/url"
	"sync"
	"time"

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

func runNewConn(wg *sync.WaitGroup) {
	defer wg.Done()

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			c.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func main() {
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go runNewConn(wg)
	}
	wg.Wait()
}
