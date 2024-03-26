package main

import (
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet/v2"
)

func NewEventHandler() *EventHandler {
	return &EventHandler{
		conns: cmap.NewStringer[Conntion, *Session](),
	}
}

type EventHandler struct {
	*gnet.BuiltinEventEngine

	conns cmap.ConcurrentMap[Conntion, *Session]
}

func (h *EventHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *EventHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	h.conns.Set(Conntion(conn.Fd()), NewSession(conn))
	return nil, gnet.None
}

func (h *EventHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	h.conns.Remove(Conntion(c.Fd()))
	return gnet.Close
}

func (h *EventHandler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	_, _ = conn.Next(-1)
	return gnet.None
}

func (h *EventHandler) OnTick() (delay time.Duration, action gnet.Action) {
	h.conns.IterCb(func(_ Conntion, v *Session) {
		v.OnTick()
	})

	return time.Second, gnet.None
}
