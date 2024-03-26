package main

import (
	"bytes"
	"log"
	"strconv"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Conntion int

func (c Conntion) String() string {
	return strconv.Itoa(int(c))
}

func NewSession(conn gnet.Conn) *Session {
	return &Session{
		conn:       conn,
		buf:        bytes.NewBuffer(make([]byte, bufferSize)),
		activeTime: time.Now().UnixMilli(),
	}
}

type Session struct {
	conn       gnet.Conn
	buf        *bytes.Buffer
	activeTime int64
}

func (s *Session) OnTick() {
	now := time.Now().UnixMilli()
	if now-s.activeTime < 5000 {
		return
	}

	s.buf.Reset()
	s.buf.WriteString("hello")
	if _, err := s.conn.Write(s.buf.Bytes()); err != nil {
		log.Printf("Error writing to server: %v\n", err.Error())
	}

	s.activeTime = now
}
