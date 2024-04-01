package main

import (
	"crypto/sha1"
	"flag"
	"log"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

var (
	addr = ""
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:12345", "kcp service address")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	// exchange pass/salt via tcp
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	listener, err := kcp.ListenWithOptions(addr, block, 10, 3)
	if err != nil {
		log.Fatal(err)
	}

	for {
		s, err := listener.AcceptKCP()
		if err != nil {
			log.Fatal(err)
		}

		go handleEcho(s)
	}
}

// handleEcho send back everything it received
func handleEcho(conn *kcp.UDPSession) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}
