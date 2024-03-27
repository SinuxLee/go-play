package main

import (
	"errors"
	"log"
	"time"

	"play/pkg/codec"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet/v2"
)

func NewEventHandler() (*EventHandler, error) {
	h := &EventHandler{
		sessions: cmap.NewStringer[Conntion, *Session](),
		coder:    &codec.SimpleCodec{},
	}

	cli, err := gnet.NewClient(
		h,
		gnet.WithTicker(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithLockOSThread(true),
		gnet.WithMulticore(true),
	)

	if err != nil {
		return nil, err
	}

	h.cli = cli
	return h, h.Start()
}

type EventHandler struct {
	*gnet.BuiltinEventEngine
	cli *gnet.Client

	sessions cmap.ConcurrentMap[Conntion, *Session]
	coder    *codec.SimpleCodec
}

func (h *EventHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *EventHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	h.sessions.Set(Conntion(conn.Fd()), NewSession(conn, h))
	return nil, gnet.None
}

func (h *EventHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	h.sessions.Remove(Conntion(c.Fd()))
	if err != nil {
		log.Printf("closed %v\n", err.Error())
	}
	return gnet.Close
}

func (h *EventHandler) OnTraffic(conn gnet.Conn) gnet.Action {
	if s, ok := h.sessions.Get(Conntion(conn.Fd())); ok {
		data, err := h.coder.Decode(s.conn)
		if errors.Is(err, codec.ErrIncompletePacket) {
			return gnet.None
		} else if err != nil {
			log.Printf("can't decode err: %v", err.Error())
			return gnet.Close
		}

		if err = s.OnData(data); err != nil {
			log.Printf("can't handle body err: %v", err.Error())
			return gnet.Close
		}
	}

	return gnet.None
}

func (h *EventHandler) OnTick() (delay time.Duration, action gnet.Action) {
	h.sessions.IterCb(func(_ Conntion, v *Session) {
		v.OnTick()
	})

	return time.Millisecond * 10, gnet.None
}

func (h *EventHandler) Send(conn gnet.Conn, data []byte) error {
	buf, err := h.coder.Encode(data)
	if err != nil {
		return err
	}

	return conn.AsyncWrite(buf, func(c gnet.Conn, err error) error {
		if err != nil {
			c.Close()
		}
		return nil
	})
}

func (h *EventHandler) Start() error {
	return h.cli.Start()
}

func (h *EventHandler) Connect(addr string) (gnet.Conn, error) {
	return h.cli.Dial("tcp", addr)
}
