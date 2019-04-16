package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

type Record struct {
	Ticker string
	Date   uint64
	Time   uint64
	Last   float64
	Volume uint64
}

func (srv *exServer) readFile(path string, out chan Record) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(file)
	var flag = false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if !flag {
			flag = true
			continue
		}
		parts := strings.Split(line, ";")
		var ticker = parts[0]
		var date, _ = strconv.ParseUint(parts[2], 10, 64)
		var time, _ = strconv.ParseUint(parts[3], 10, 64)
		var last, _ = strconv.ParseFloat(parts[4], 64)
		var volume, _ = strconv.ParseUint(parts[5], 10, 64)
		currentRecord := Record{
			Ticker: ticker,
			Date:   date,
			Time:   time,
			Last:   last,
			Volume: volume,
		}
		//fmt.Printf("READFILE: %v", currentRecord)
		out <- currentRecord
		srv.CheckDeals(currentRecord)
	}
	close(out)
	return nil
}

func (srv *exServer) CheckDeals(rec Record) {
	//srv.completed.Add()
	indexForDelete := []int{}
	srv.storage.Mu.Lock()
	for i, val := range srv.storage.SellRequests {
		if val.Ticker == rec.Ticker && val.Price <= float32(rec.Last) {
			srv.completed.Add(val)
			indexForDelete = append(indexForDelete, i)
		}
	}

	for i := len(indexForDelete) - 1; i >= 0; i-- {
		srv.storage.SellRequests = append(srv.storage.SellRequests[:indexForDelete[i]], srv.storage.SellRequests[indexForDelete[i+1]:]...)
	}

	indexForDelete = []int{}

	for i, val := range srv.storage.BuyRequests {
		if val.Ticker == rec.Ticker && val.Price >= float32(rec.Last) {
			if val.Amount <= int32(rec.Volume) { //продаем всё
				srv.completed.Add(val)
				indexForDelete = append(indexForDelete, i)
			} else if rec.Volume != 0 { //покупаем часть
				val.Partial = true
				val.Amount = int32(rec.Volume)
				srv.completed.Add(val)
				srv.storage.BuyRequests[i].Amount -= int32(rec.Volume)
			}
		}
	}
	for i := len(indexForDelete) - 1; i >= 0; i-- {
		srv.storage.BuyRequests = append(srv.storage.BuyRequests[:indexForDelete[i]], srv.storage.BuyRequests[indexForDelete[i+1]:]...)
	}
	srv.storage.Mu.Unlock()
}
