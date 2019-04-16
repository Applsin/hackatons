package main

import (
	"context"
	"log"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const listenAddr string = "127.0.0.1:8082"

func wait(amout int) {
	time.Sleep(time.Duration(amout) * 10 * time.Millisecond)
}

func getGrpcConn(t *testing.T) *grpc.ClientConn {
	grcpConn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("cant connect to grpc: %v", err)
	}
	return grcpConn
}

func TestServerStartStop(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())
	StartMicroservice(ctx, listenAddr)
	wait(1)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт
	wait(1)

	// теперь проверим что вы освободили порт и мы можем стартовать сервер ещё раз
	ctx, finish = context.WithCancel(context.Background())
	StartMicroservice(ctx, listenAddr)
	wait(1)
	finish()
	wait(1)
}

func TestServerAll(t *testing.T) {
	wait(1)
	ctx, finish := context.WithCancel(context.Background())
	StartMicroservice(ctx, listenAddr)
	defer func() {
		finish()
		wait(1)
	}()
	conn := getGrpcConn(t)
	defer conn.Close()
	broker := NewExchangeClient(conn)
	deal := Deal{
		ID:       5,
		BrokerID: 10,
		ClientID: 100500,
		Ticker:   "IMOEX",
		Amount:   10,
		Partial:  false,
		Time:     0,
		Price:    2322.6,
	}
	md := metadata.Pairs(
		"TYPE", "SELL",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)
	var dealID, err = broker.Create(ctx, &deal)
	log.Println(dealID, err)
}
