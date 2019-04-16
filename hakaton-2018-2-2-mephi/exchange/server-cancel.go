package main

import (
	"context"
	//"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
)

func (s *exServer) Cancel(ctx context.Context, dealId *DealID) (*CancelResult, error) {
	val := Deal{}

	s.storage.Mu.Lock()
	for i, val := range s.storage.BuyRequests {
		if dealId.ID == val.ID {
			s.storage.BuyRequests = append(s.storage.BuyRequests[:i], s.storage.BuyRequests[i+1:]...)
			s.storage.Mu.Unlock()
			return &CancelResult{true, val.XXX_NoUnkeyedLiteral, val.XXX_unrecognized, val.XXX_sizecache}, nil
		}
	}
	for i, val := range s.storage.SellRequests {
		if dealId.ID == val.ID {
			s.storage.SellRequests = append(s.storage.SellRequests[:i], s.storage.SellRequests[i+1:]...)
			s.storage.Mu.Unlock()
			return &CancelResult{true, val.XXX_NoUnkeyedLiteral, val.XXX_unrecognized, val.XXX_sizecache}, nil
		}
	}
	s.storage.Mu.Unlock()
	return &CancelResult{false, val.XXX_NoUnkeyedLiteral, val.XXX_unrecognized, val.XXX_sizecache}, nil
	//return nil, grpc.Errorf(codes.Unavailable, "no such deal")
}
