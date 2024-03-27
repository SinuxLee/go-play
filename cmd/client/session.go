package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"play/pkg/schema"

	"github.com/panjf2000/gnet/v2"
)

type Conntion int

func (c Conntion) String() string {
	return strconv.Itoa(int(c))
}

func NewSession(conn gnet.Conn, h *EventHandler) *Session {
	p := &schema.Player{
		UserId:         uint64(rand.Int63()),
		Avatar:         uint16(rand.Int31()),
		AvatarURL:      fmt.Sprintf("https://ffa-game1.diandian.info/%v/avatar", rand.Int()),
		AvatarFrame:    uint16(rand.Int31()),
		Level:          uint16(rand.Int31()),
		GuildId:        uint32(rand.Int31()),
		GuildName:      fmt.Sprintf("https://ffa-game1.diandian.info/%v/avatar", rand.Int()),
		GuildIcon:      uint16(rand.Int31()),
		GuildIconFrame: uint16(rand.Int31()),
		GuildIconBG:    uint32(rand.Int31()),
		Nick:           fmt.Sprintf("https://ffa-game1.diandian.info/%v/avatar", rand.Int()),
		Likes:          uint32(rand.Int31()),
		GoldBadge:      uint32(rand.Int31()),
		SilverBadge:    uint32(rand.Int31()),
		BronzeBadge:    uint32(rand.Int31()),
		CollectLevel:   uint32(rand.Int31()),
	}

	return &Session{
		handler:    h,
		conn:       conn,
		activeTime: time.Now().UnixMilli(),
		player:     p,
	}
}

type Session struct {
	handler    *EventHandler
	conn       gnet.Conn
	activeTime int64
	player     *schema.Player
}

func (s *Session) OnData(data []byte) error {
	m := make(map[string]any)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// log.Printf("%+v\n", len(data))
	return nil
}

func (s *Session) OnTick() {
	now := time.Now().UnixMilli()
	if now-s.activeTime < 100 {
		return
	}

	data, _ := json.Marshal(s.player)
	if err := s.handler.Send(s.conn, data); err != nil {
		log.Printf("Error writing to server: %v\n", err.Error())
	}

	s.activeTime = now
}
