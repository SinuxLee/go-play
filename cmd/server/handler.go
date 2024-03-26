package main

import (
	"log"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet/v2"
)

const heartbeatInterval = 3

func NewEventHandler() *EventHandler {
	return &EventHandler{
		sessions: cmap.NewStringer[Conntion, *Session](),
	}
}

type EventHandler struct {
	*gnet.BuiltinEventEngine

	eng      gnet.Engine
	sessions cmap.ConcurrentMap[Conntion, *Session]
}

func (h *EventHandler) Run(addr string) {
	gnet.Run(h, addr,
		gnet.WithMulticore(true),
		gnet.WithLockOSThread(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithTicker(true),
	)
}

func (h *EventHandler) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	h.sessions.Set(Conntion(c.Fd()), NewSession(c))

	return nil, gnet.None
}

func (h *EventHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	h.sessions.Remove(Conntion(c.Fd()))
	return gnet.Close
}

func (h *EventHandler) OnBoot(eng gnet.Engine) gnet.Action {
	h.eng = eng
	return gnet.None
}

func (h *EventHandler) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	c.AsyncWrite(buf, nil)

	if s, ok := h.sessions.Get(Conntion(c.Fd())); ok {
		s.activeTime = time.Now().Unix()
	}

	return gnet.None
}

func (h *EventHandler) OnTick() (delay time.Duration, action gnet.Action) {
	now := time.Now().Unix()
	offline := make([]Conntion, 0, 64)

	h.sessions.IterCb(func(key Conntion, v *Session) {
		if now-v.activeTime > heartbeatInterval*3 {
			v.conn.Close()
			offline = append(offline, key)
			return
		}

		if now-v.lastBeatTime > heartbeatInterval {
			v.conn.AsyncWrite([]byte("heartbeat\n"), nil)
			v.lastBeatTime = now
		}
	})

	for _, v := range offline {
		h.sessions.Remove(v)
	}

	log.Printf("clients %v\n", h.sessions.Count())
	return time.Second, gnet.None
}
