package main

import (
	"crypto/sha1"
	"flag"
	"io"
	"log"
	"time"

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
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)

	sess, err := kcp.DialWithOptions(addr, block, 10, 3)
	if err != nil {
		log.Fatal(err)
	}

	for {
		data := time.Now().String()
		buf := make([]byte, len(data))

		if _, err := sess.Write([]byte(data)); err != nil {
			log.Fatal(err)
		}

		_, err := io.ReadFull(sess, buf)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("recv:", string(buf))
		time.Sleep(time.Second)
	}
}
