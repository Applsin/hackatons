package main

// exServer implements ExchangeServer
type exServer struct {
	storage   *Storage
	records   []chan Record
	completed ConcurrentQueue
}

func NewExchangeServer(s *Storage, c []chan Record) exServer {
	return exServer{storage: s, records: c}
}
