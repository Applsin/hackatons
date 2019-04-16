package main

import (
	"fmt"
	"time"
)

func (s *exServer) Results(brokerId *BrokerID, ers Exchange_ResultsServer) error {
	for {
		res := s.completed.Get(*brokerId)
		for _, j := range res {
			fmt.Printf("RESULT %v\n", j)
			ers.Send(&j)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
