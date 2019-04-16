package main

import (
	"sync"
)

type ConcurrentQueue struct {
	mu    sync.Mutex
	deals []Deal
}

func (q *ConcurrentQueue) Add(deal Deal) {
	q.mu.Lock()
	q.deals = append(q.deals, deal)
	q.mu.Unlock()
}

func (q *ConcurrentQueue) Get(b BrokerID) []Deal {
	q.mu.Lock()
	var temp = []Deal{}
	for i, item := range q.deals {
		if int64(item.BrokerID) == b.ID {
			temp = append(temp, item)
			q.deals = append(q.deals[:i], q.deals[i+1:]...)
		}
	}
	q.mu.Unlock()
	return temp
}
