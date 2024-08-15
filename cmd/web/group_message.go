package main

import "sync"

type msgGroup struct {
	messageBuffer int
	mu            sync.Mutex
	activeConns   map[*msgSubscriber]struct{}
}
