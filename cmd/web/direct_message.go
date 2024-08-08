package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Tsundere-Musume/message/internal/models"
	"github.com/julienschmidt/httprouter"
	"nhooyr.io/websocket"
)

type directMsgServer struct {
	room map[string]*dmRoom
	mu   sync.Mutex
}

type dmRoom struct {
	messageBuffer int
	mu            sync.Mutex
	activeConns   map[*msgSubscriber]struct{}
}

type msgSubscriber struct {
	msgs      chan models.DirectMessage
	closeSlow func()
}

func serverDM() *directMsgServer {
	return &directMsgServer{
		room: make(map[string]*dmRoom),
	}
}

func newDMRoom() *dmRoom {
	return &dmRoom{
		messageBuffer: 16,
		activeConns:   make(map[*msgSubscriber]struct{}),
	}
}

func (s *directMsgServer) getRoomByIds(id1, id2 string) (*dmRoom, bool) {
	key1 := fmt.Sprintf("%s:%s", id1, id2)
	var room *dmRoom
	var ok bool
	s.mu.Lock()
	defer s.mu.Unlock()
	if room, ok = s.room[key1]; ok {
		return room, true
	}
	delete(s.room, key1)
	key2 := fmt.Sprintf("%s:%s", id2, id1)
	if room, ok = s.room[key2]; ok {
		return room, true
	}
	delete(s.room, key1)
	return nil, false
}

func (s *directMsgServer) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	senderId, ok := r.Context().Value(userIdKey).(string)
	if !ok || senderId == "" {
		return errors.New("Invalid sender id")
	}
	//TODO: validate id
	params := httprouter.ParamsFromContext(r.Context())
	receiverId := params.ByName("id")
	if receiverId == "" {
		return models.ErrNotFound
	}
	room, ok := s.getRoomByIds(senderId, receiverId)
	if !ok {
		key := fmt.Sprintf("%s:%s", senderId, receiverId)
		s.mu.Lock()
		s.room[key] = newDMRoom()
		room = s.room[key]
		s.mu.Unlock()
	}
	return room.subscribe(ctx, w, r)
}

func (room *dmRoom) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	sub := &msgSubscriber{
		msgs: make(chan models.DirectMessage, room.messageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages.")
			}
		},
	}

	room.addSubscriber(sub)
	defer room.deleteSubscriber(sub)

	c2, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}

	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}

	c = c2
	mu.Unlock()
	defer c.CloseNow()
	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-sub.msgs:
			err := writeTimeouts(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func (room *dmRoom) addSubscriber(s *msgSubscriber) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.activeConns[s] = struct{}{}
}

func (room *dmRoom) deleteSubscriber(s *msgSubscriber) {
	room.mu.Lock()
	defer room.mu.Unlock()
	delete(room.activeConns, s)
}

func (room *dmRoom) publish(msg models.DirectMessage) {
	room.mu.Lock()
	defer room.mu.Unlock()

	// TODO: rate limiter

	for s := range room.activeConns {
		select {
		case s.msgs <- msg:

		default:
			go s.closeSlow()
		}
	}
}

func writeTimeouts(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg models.DirectMessage) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	m, err := msg.Serialize()
	if err != nil {
		return err
	}

	return c.Write(ctx, websocket.MessageText, m)
}
