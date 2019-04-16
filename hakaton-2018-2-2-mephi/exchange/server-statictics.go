package main

import (
	"fmt"
	"sync"
	"time"
)

func (s *exServer) Statistic(brokerId *BrokerID, es Exchange_StatisticServer) error {
	wg := &sync.WaitGroup{}
	fmt.Println("Stats start")
	for i, j := range s.records {
		//fmt.Printf("Iterating chan %d\n", i)
		wg.Add(1)
		go func(i int, j chan Record) {
			defer wg.Done()
			var cnt int64
			Prev := <-j
			collection := make([]Record, 0)
			collection = append(collection, Prev)
			for Next := range j {
				if Prev.Time == Next.Time {
					collection = append(collection, Next)
					//fmt.Printf("[%d] equal next=%v , prev=%v\n", i, Next.Time, Prev.Time)
				} else {
					//fmt.Printf("[%d] not next=%v , prev=%v\n", i, Next, Prev)
					collection = append(collection, Prev)
					Prev = Next
					//fmt.Printf("collection len=%v\n", len(collection))
					//	collection = collection[:len(collection)]

					init := collection[0]
					var volume int64
					open, high, low := init.Last, init.Last, init.Last
					for i := range collection {
						if collection[i].Last > high {
							high = collection[i].Last
						}
						if collection[i].Last < low {
							low = collection[i].Last
						}
						volume += int64(collection[i].Volume)
					}
					close := collection[len(collection)-1].Last
					es.Send(&OHLCV{
						ID:       cnt,
						Time:     int32(time.Now().Unix()),
						Interval: 1,
						Open:     float32(open),
						High:     float32(high),
						Low:      float32(low),
						Close:    float32(close),
						Ticker:   collection[0].Ticker,
						Volume:   volume,
					})
					cnt++
					time.Sleep(1 * time.Second)
					collection := make([]Record, 0)
					collection = append(collection, Prev)
				}
			}
		}(i, j)
	}
	wg.Wait()
	return nil
}

//OHLCV = uint32(time.Now().Unix)
