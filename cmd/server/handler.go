package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"play/internal/codec"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/buffer/ring"
	"go.uber.org/atomic"
)

const heartbeatInterval = 3

type MessageMaker interface {
	Process(s *Session, data []byte) error
}

type MakeMessageFun func(s *Session, data []byte) error

func (f MakeMessageFun) Process(s *Session, data []byte) error {
	return f(s, data)
}

func NewEventHandler(c codec.Coder, m MessageMaker) *EventHandler {
	return &EventHandler{
		sessions: cmap.NewStringer[Conntion, *Session](),
		coder:    c,
		maker:    m,
	}
}

type EventHandler struct {
	*gnet.BuiltinEventEngine

	eng      gnet.Engine
	sessions cmap.ConcurrentMap[Conntion, *Session]
	coder    codec.Coder
	maker    MessageMaker

	recvCounter atomic.Uint32
}

func (h *EventHandler) Run(addr string) {
	gnet.Run(h, addr,
		gnet.WithMulticore(true),
		gnet.WithLockOSThread(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithTicker(true),
		gnet.WithReadBufferCap(8*ring.DefaultBufferSize),
		gnet.WithReadBufferCap(8*ring.DefaultBufferSize),
	)
}

func (h *EventHandler) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	h.sessions.Set(Conntion(c.Fd()), NewSession(c, h))
	return nil, gnet.None
}

func (h *EventHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	h.sessions.Remove(Conntion(c.Fd()))
	if err != nil {
		log.Printf("%v closed, %v\n", c.Fd(), err.Error())
	}

	return gnet.Close
}

func (h *EventHandler) OnBoot(eng gnet.Engine) gnet.Action {
	h.eng = eng
	return gnet.None
}

func (h *EventHandler) OnTraffic(c gnet.Conn) gnet.Action {
	s, ok := h.sessions.Get(Conntion(c.Fd()))
	if !ok {
		h.sessions.Remove(Conntion(c.Fd()))
		return gnet.Close
	}

	// 循环读取 buffer 中的消息
	for {
		data, err := h.coder.Decode(s.conn)
		if errors.Is(err, codec.ErrIncompletePacket) ||
			errors.Is(err, io.ErrShortBuffer) {
			return gnet.None
		} else if err != nil {
			h.sessions.Remove(Conntion(c.Fd()))
			log.Printf("%v can't decode buf, err: %v\n", c.Fd(), err.Error())
			return gnet.Close
		}

		// Discard message
		if h.maker == nil {
			return gnet.None
		}

		if err = h.maker.Process(s, data); err != nil {
			h.sessions.Remove(Conntion(c.Fd()))
			log.Printf("%v can't handle body, err: %v\n", c.Fd(), err.Error())
			return gnet.Close
		}
		h.recvCounter.Add(1)
		s.activeTime = time.Now().Unix()
	}
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

	log.Printf("clients %v, recv qps: %v \n", h.sessions.Count(), h.recvCounter.Load())

	h.recvCounter.Store(0)
	return time.Second, gnet.None
}

func (h *EventHandler) Send(conn gnet.Conn, data []byte) error {
	buf, err := h.coder.Encode(data)
	if err != nil {
		return err
	}

	return conn.AsyncWrite(buf, func(c gnet.Conn, err error) error {
		if err != nil {
			h.sessions.Remove(Conntion(c.Fd()))
			c.Close()
		}
		return nil
	})
}
