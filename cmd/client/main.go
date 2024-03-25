package main

import (
	"bytes"
	"flag"
	"log"
	"strconv"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet/v2"
)

const (
	numClients = 50000
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

type EventHandler struct {
	*gnet.BuiltinEventEngine

	conns cmap.ConcurrentMap[Conntion, *session]
}

func (h *EventHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	h.conns.Set(Conntion(conn.Fd()), newSession(conn))
	return nil, gnet.None
}

func (h *EventHandler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	_, _ = conn.Next(-1)
	return gnet.None
}

func (h *EventHandler) OnTick() (delay time.Duration, action gnet.Action) {
	now := time.Now().UnixMilli()
	h.conns.IterCb(func(_ Conntion, v *session) {
		if now-v.activeTime < 1000 {
			return
		}

		v.buf.Reset()
		v.buf.WriteString("hello")
		if _, err := v.conn.Write(v.buf.Bytes()); err != nil {
			log.Printf("Error writing to server: %v\n", err.Error())
		}

		v.activeTime = now
	})

	return time.Second, gnet.None
}

func newSession(conn gnet.Conn) *session {
	return &session{
		conn:       conn,
		buf:        bytes.NewBuffer(make([]byte, bufferSize)),
		activeTime: time.Now().UnixMilli(),
	}
}

type session struct {
	conn       gnet.Conn
	buf        *bytes.Buffer
	activeTime int64
}

type Conntion int

func (c Conntion) String() string {
	return strconv.Itoa(int(c))
}

func main() {
	cli, err := gnet.NewClient(
		&EventHandler{
			conns: cmap.NewStringer[Conntion, *session](),
		},
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
