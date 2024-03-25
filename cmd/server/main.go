package main

import (
	"log"
	"time"

	"github.com/panjf2000/gnet/v2"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

type echoServer struct {
	*gnet.BuiltinEventEngine

	eng gnet.Engine
}

func (es *echoServer) OnBoot(eng gnet.Engine) gnet.Action {
	es.eng = eng
	return gnet.None
}

func (es *echoServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	c.Write(buf)
	return gnet.None
}

func (es *echoServer) OnTick() (delay time.Duration, action gnet.Action) {
	return time.Second, gnet.None
}

func main() {
	echo := &echoServer{}
	gnet.Run(echo, "tcp://:8888",
		gnet.WithMulticore(true),
		gnet.WithLockOSThread(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithTicker(true),
	)
}
