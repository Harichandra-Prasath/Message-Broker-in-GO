package main

import (
	"fmt"
	"sync"
)

type Storer interface {
	Append([]byte) int
	Pop(int) ([]byte, error)
}

type ProduceFunc func() Storer

type Store struct {
	data [][]byte
	mu   sync.RWMutex
}

func produce() Storer {
	return NewStore()
}
func NewStore() *Store {
	return &Store{
		data: make([][]byte, 0),
	}
}

func (st *Store) Append(data []byte) int {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.data = append(st.data, data)
	return len(st.data) - 1
}

func (st *Store) Pop(pos int) ([]byte, error) {
	//check for position errors
	if pos <= 0 {
		return nil, fmt.Errorf("invalid Position to pop")
	} else if pos > len(st.data) {
		return nil, fmt.Errorf("%d Position too high...Should be less than %d", pos, len(st.data)+1)
	}
	return st.data[pos-1], nil
}
