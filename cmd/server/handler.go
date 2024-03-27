package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"play/pkg/codec"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet/v2"
)

const heartbeatInterval = 3

func NewEventHandler() *EventHandler {
	return &EventHandler{
		sessions: cmap.NewStringer[Conntion, *Session](),
		coder:    &codec.SimpleCodec{},
	}
}

type EventHandler struct {
	*gnet.BuiltinEventEngine

	eng      gnet.Engine
	sessions cmap.ConcurrentMap[Conntion, *Session]
	coder    *codec.SimpleCodec
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
	h.sessions.Set(Conntion(c.Fd()), NewSession(c, h))

	return nil, gnet.None
}

func (h *EventHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	h.sessions.Remove(Conntion(c.Fd()))
	if err != nil {
		log.Printf("closed %v\n", err.Error())
	}

	return gnet.Close
}

func (h *EventHandler) OnBoot(eng gnet.Engine) gnet.Action {
	h.eng = eng
	return gnet.None
}

func (h *EventHandler) OnTraffic(c gnet.Conn) gnet.Action {
	if s, ok := h.sessions.Get(Conntion(c.Fd())); ok {
		s.activeTime = time.Now().Unix()

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
	now := time.Now().Unix()
	offline := make([]Conntion, 0, 64)

	h.sessions.IterCb(func(key Conntion, v *Session) {
		if now-v.activeTime > heartbeatInterval*3 {
			log.Printf("inactive, close %v\n", key)
			v.conn.Close()
			offline = append(offline, key)
			return
		}

		if now-v.lastBeatTime > heartbeatInterval {
			data, _ := json.Marshal(map[string]any{
				"event": "heartbeat",
				"time":  now / 1000,
			})

			if buf, err := h.coder.Encode(data); err == nil {
				v.conn.AsyncWrite(buf, nil)
				v.lastBeatTime = now
			}
		}
	})

	for _, v := range offline {
		h.sessions.Remove(v)
	}

	log.Printf("clients %v\n", h.sessions.Count())
	return time.Second, gnet.None
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
