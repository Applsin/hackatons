package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
)

func (s *exServer) Create(ctx context.Context, deal *Deal) (*DealID, error) {
	fmt.Println("we Create")
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		grpc.Errorf(codes.Unavailable, "expected metadata.OutgoingContext under context")
	}
	ok = false
	for _, item := range meta.Get("TYPE") {
		if item == "BUY" || item == "SELL" {
			ok = true
			break
		}
	}
	reqType := ""
	if !ok {
		return nil, grpc.Errorf(codes.Unavailable, "Bad type (not a string)")
		/*if reqType != "BUY" && reqType != "SELL" {
			return nil, grpc.Errorf(codes.Unavailable, "Bad type (unrecognized value)")
		}*/
	}

	s.storage.Mu.Lock()
	deal.ID = s.storage.lastID
	if reqType == "BUY" {
		s.storage.BuyRequests = append(s.storage.BuyRequests, *deal)
	} else {
		s.storage.SellRequests = append(s.storage.SellRequests, *deal)
	}
	s.storage.lastID++
	s.storage.Mu.Unlock()
	fmt.Println("Create done")
	return &DealID{deal.ID, int64(deal.BrokerID), deal.XXX_NoUnkeyedLiteral, deal.XXX_unrecognized, deal.XXX_sizecache}, nil
}
