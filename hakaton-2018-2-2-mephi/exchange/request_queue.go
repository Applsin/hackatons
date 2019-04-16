package main

import "sync"

type Storage struct {
	lastID       int64
	SellRequests []Deal
	BuyRequests  []Deal
	Mu           *sync.Mutex
}
