package main

//
// import (
// 	"context"
// 	"net"
// 	"net/http"
// 	"sync"
// 	"time"
//
// 	"nhooyr.io/websocket"
// )
//
// type chatRoom struct {
// 	subscriberMessageBuffer int
// 	subsriberMu             sync.Mutex
// 	subscribers             map[*subscriber]struct{}
// }
//
// type message struct {
// 	UserID   string `json:"userId"`
// 	Username string `json:"username"`
// 	Value    string `json:"value"`
// }
//
// type subscriber struct {
// 	msgs      chan message
// 	closeSlow func()
// }
//
// func newChatServer() *chatRoom {
// 	return &chatRoom{
// 		subscriberMessageBuffer: 16,
// 		subscribers:             make(map[*subscriber]struct{}),
// 	}
// }
//
// // TODO: understand this
// func (croom *chatRoom) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
// 	var mu sync.Mutex
// 	var c *websocket.Conn
// 	var closed bool
// 	s := &subscriber{
// 		msgs: make(chan message, croom.subscriberMessageBuffer),
// 		closeSlow: func() {
// 			mu.Lock()
// 			defer mu.Unlock()
// 			closed = true
// 			if c != nil {
// 				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages.")
// 			}
// 		},
// 	}
//
// 	croom.addSubscriber(s)
// 	defer croom.deleteSubscriber(s)
// 	c2, err := websocket.Accept(w, r, nil)
// 	if err != nil {
// 		return err
// 	}
//
// 	mu.Lock()
// 	if closed {
// 		mu.Unlock()
// 		return net.ErrClosed
// 	}
//
// 	c = c2
// 	mu.Unlock()
// 	defer c.CloseNow()
//
// 	ctx = c.CloseRead(ctx)
//
// 	for {
// 		select {
// 		case msg := <-s.msgs:
// 			err := writeTimeout(ctx, time.Second*5, c, msg)
// 			if err != nil {
// 				return err
// 			}
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		}
// 	}
// }
//
// func (croom *chatRoom) addSubscriber(s *subscriber) {
// 	croom.subsriberMu.Lock()
// 	defer croom.subsriberMu.Unlock()
//
// 	croom.subscribers[s] = struct{}{}
// }
//
// func (croom *chatRoom) deleteSubscriber(s *subscriber) {
// 	croom.subsriberMu.Lock()
// 	defer croom.subsriberMu.Unlock()
//
// 	delete(croom.subscribers, s)
//
// }
//
// func (croom *chatRoom) publish(msg message) {
// 	croom.subsriberMu.Lock()
// 	defer croom.subsriberMu.Unlock()
//
// 	// TODO: rate limiter
//
// 	for s := range croom.subscribers {
// 		select {
// 		case s.msgs <- msg:
// 		default:
// 			go s.closeSlow()
// 		}
// 	}
// }
//
// func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg message) error {
// 	ctx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()
//
// 	m, err := serializeMessage(msg)
// 	if err != nil {
// 		return err
// 	}
//
// 	return c.Write(ctx, websocket.MessageText, m)
// }
