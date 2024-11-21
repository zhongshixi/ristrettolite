package ristrettolite

import (
	"sync"
	"time"
)

type Action int

const (
	ActionNone Action = iota
	ActionPut
	ActionRemove
	ActionWait
)

type Item[T any] struct {
	// core data
	Key      uint64
	Value    T
	ExpireAt time.Time
	Cost     int
	Action   Action

	WaitGroup *sync.WaitGroup

	//  meta data
	// index of the node in the priority queue
	pqIndex int
}
