package main

import (
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

type MessageMaker interface {
	Process(s *Session, data []byte) error
}

type MakeMessageFun func(s *Session, data []byte) error

func (f MakeMessageFun) Process(s *Session, data []byte) error {
	return f(s, data)
}

func NewEventHandler(c codec.Coder, m MessageMaker) (*EventHandler, error) {
	h := &EventHandler{
		sessions: cmap.NewStringer[Conntion, *Session](),
		coder:    c,
		maker:    m,
	}

	cli, err := gnet.NewClient(
		h,
		gnet.WithTicker(true),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithLockOSThread(true),
		gnet.WithMulticore(true),
		gnet.WithReadBufferCap(8*ring.DefaultBufferSize),
		gnet.WithReadBufferCap(8*ring.DefaultBufferSize),
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
	coder    codec.Coder
	maker    MessageMaker

	recvCounter atomic.Uint32
	sendCounter atomic.Uint32
}

func (h *EventHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *EventHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	h.sessions.Set(Conntion(conn.Fd()), NewSession(conn, h))
	return nil, gnet.None
}

func (h *EventHandler) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	h.sessions.Remove(Conntion(conn.Fd()))
	if err != nil {
		log.Printf("closed %v\n", err.Error())
	}
	return gnet.Close
}

func (h *EventHandler) OnTraffic(conn gnet.Conn) gnet.Action {
	s, ok := h.sessions.Get(Conntion(conn.Fd()))
	if !ok {
		h.sessions.Remove(Conntion(conn.Fd()))
		return gnet.Close
	}

	for {
		data, err := h.coder.Decode(s.conn)
		if errors.Is(err, codec.ErrIncompletePacket) ||
			errors.Is(err, io.ErrShortBuffer) {
			return gnet.None
		} else if err != nil {
			log.Printf("can't decode err: %v", err.Error())
			h.sessions.Remove(Conntion(conn.Fd()))
			return gnet.Close
		}

		// Discard message
		if h.maker == nil {
			continue
		}

		if err = h.maker.Process(s, data); err != nil {
			log.Printf("can't handle body err: %v", err.Error())
			h.sessions.Remove(Conntion(conn.Fd()))
			return gnet.Close
		}

		h.recvCounter.Add(1)
	}
}

func (h *EventHandler) OnTick() (delay time.Duration, action gnet.Action) {
	h.sessions.IterCb(func(_ Conntion, v *Session) {
		v.OnTick()
		h.sendCounter.Add(1)
	})

	log.Printf("send qps %v, recv %v", h.sendCounter.Load(), h.recvCounter.Load())
	h.recvCounter.Store(0)
	h.sendCounter.Store(0)

	return time.Second, gnet.None
}

func (h *EventHandler) Send(conn gnet.Conn, data []byte) error {
	buf, err := h.coder.Encode(data)
	if err != nil {
		return err
	}

	return conn.AsyncWrite(buf, func(c gnet.Conn, err error) error {
		if err != nil {
			h.sessions.Remove(Conntion(conn.Fd()))
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
