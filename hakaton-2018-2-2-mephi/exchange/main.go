package main

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

func main() {
	// нужен listen addr от брокера
	listenAddr := "127.0.0.1:8082"
	fmt.Println("Starting program")
	StartMicroservice(context.Background(), listenAddr)
}

func StartMicroservice(ctx context.Context, listenAddr string) {
	//go func() {
	server := grpc.NewServer()
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return
	}
	storage := Storage{
		Mu:           &sync.Mutex{},
		SellRequests: make([]Deal, 0),
		BuyRequests:  make([]Deal, 0),
	}
	chans := []chan Record{
		make(chan Record),
		make(chan Record),
		make(chan Record),
	}
	exchange := NewExchangeServer(&storage, chans)
	RegisterExchangeServer(server, &exchange)
	paths := []string{
		"./IMOEX_180518_180518.txt",
		"./SPFB.RTS_180518_180518.txt",
		"./USD000UTSTOM_180518_180518.txt",
	}
	for i := 0; i < len(paths); i++ {
		go exchange.readFile(paths[i], chans[i]) // need WaitGroup
	}
	//go func() {
	fmt.Println("Starting serve")
	server.Serve(lis)
	//}()
	// select {
	// case <-ctx.Done():
	// 	server.Stop()
	// }
	//}()
}
