package main

import (
	"strconv"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Conntion int

func (c Conntion) String() string {
	return strconv.Itoa(int(c))
}

func NewSession(c gnet.Conn) *Session {
	now := time.Now().Unix()
	return &Session{
		conn:         c,
		activeTime:   now,
		lastBeatTime: now,
	}
}

type Session struct {
	conn         gnet.Conn
	lastBeatTime int64 // 上一个心跳发送时间
	activeTime   int64 // 玩家最后活跃时间
}
