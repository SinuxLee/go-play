package main

import (
	"encoding/json"
	"play/pkg/schema"
	"strconv"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Conntion int

func (c Conntion) String() string {
	return strconv.Itoa(int(c))
}

func NewSession(c gnet.Conn, h *EventHandler) *Session {
	now := time.Now().Unix()
	return &Session{
		handler:      h,
		conn:         c,
		activeTime:   now,
		lastBeatTime: now,
	}
}

type Session struct {
	handler      *EventHandler
	conn         gnet.Conn
	lastBeatTime int64 // 上一个心跳发送时间
	activeTime   int64 // 玩家最后活跃时间
}

func (s *Session) OnData(data []byte) error {
	p := &schema.Player{}
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}

	return s.handler.Send(s.conn, data)
}
