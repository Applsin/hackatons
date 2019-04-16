package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Meta struct {
	HttpIp     string
	WritersNum int
	FirstPort  int
	Salt       string
}

func WorkerN(id int, salt string, n int) int {
	strID := strconv.Itoa(id) + salt
	crcH := crc32.ChecksumIEEE([]byte(strID))
	return int(crcH) % n
}

func Send(meta *Meta, in chan Record, wg *sync.WaitGroup) {
	defer wg.Done()
	for record := range in {
		jRec, err := json.Marshal(record)
		if err != nil {
			fmt.Println("Recieved bad json")
		}
		port := meta.FirstPort + WorkerN(record.Id, meta.Salt, meta.WritersNum)
		currentUrl := meta.HttpIp + strconv.Itoa(port) + "/"

		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		client := &http.Client{
			Timeout:   time.Second * 10,
			Transport: transport,
		}

		body := bytes.NewBuffer(jRec)
		req, _ := http.NewRequest(http.MethodPost, currentUrl, body)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("error in request happend", err)
			return
		}
		resp.Body.Close()
	}
}
