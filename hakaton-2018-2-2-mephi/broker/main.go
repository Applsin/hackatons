package main

import (
	"encoding/json"
	"fmt"
	"hakaton-2018-2-2-mephi/broker/bdaccess"
	"hakaton-2018-2-2-mephi/broker/config"
	"hakaton-2018-2-2-mephi/broker/info"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var (
	brokerClient info.ExchangeClient
	myBrokerID   *info.BrokerID
	currID       int64
	DB           *bdaccess.DatabaseAccessor
)

type Prices struct {
	Time   int32   `json: "time"`
	Open   float32 `json: "open"`
	High   float32 `json: "high"`
	Low    float32 `json: "low"`
	Close  float32 `json: "close"`
	Volume int64   `json: "volume"`
}

type AR map[string]interface{}

func ShowResponse(w http.ResponseWriter, okResp AR, err string, errCode int) {
	var resp []byte
	if err == "" {
		resp, _ = json.Marshal(AR{"body": okResp})
	} else {
		w.WriteHeader(errCode)
		resp, _ = json.Marshal(AR{"error": err})
	}

	w.Write(resp)
}

func ShowStatReq(w http.ResponseWriter, r *http.Request) {
	client := DB.SelectClientByLogin(r.Header.Get("X-Login"))
	if client.ID == 0 {
		ShowResponse(w, AR{}, "unauthorized", 403)
		fmt.Println(client)
		return
	}
	currAnswer := AR{}

	currAnswer["balance"] = client.Balance

	//pos := []*Position{}
	tempSlice := []AR{}

	pos := DB.SelectPositionsByClientId(client.ID)
	for _, val := range pos {
		currPos := AR{}
		currPos["ticker"] = val.Ticker
		currPos["vol"] = val.Vol
		if val.Vol > 0 {
			currPos["type"] = "BUY"
		} else {
			currPos["type"] = "SELL"
		}

		tempSlice = append(tempSlice, currPos)
	}
	currAnswer["positions"] = tempSlice
	tempSlice = []AR{}
	req := DB.SelectRequestsByClientId(client.ID)
	fmt.Println("Show Stats Req")
	//fmt.Println(req)
	for _, val := range req {
		currPos := AR{}
		currPos["ticker"] = val.Ticker
		currPos["vol"] = val.Vol
		currPos["price"] = val.Price
		currPos["status"] = val.Bought
		if val.Vol > 0 {
			currPos["type"] = "BUY"
		} else {
			currPos["type"] = "SELL"
		}
		tempSlice = append(tempSlice, currPos)
	}
	currAnswer["open_orders"] = tempSlice

	fmt.Println(currAnswer)
	ShowResponse(w, currAnswer, "", 200)
}

func ShowHistory(w http.ResponseWriter, r *http.Request) {
	client := DB.SelectClientByLogin(r.Header.Get("X-Login"))
	if client.ID == 0 {
		ShowResponse(w, AR{}, "unauthorized", 403)
		fmt.Println(client)
		return
	}

	vars := mux.Vars(r)
	time := time.Now().Unix() - 300
	stats := DB.SelectStatsByTickerAndTime(vars["ticker"], time)
	var hists []Prices
	for _, stat := range stats {
		hist := Prices{}
		hist.Time = stat.Time
		hist.Close = stat.Close
		hist.High = stat.High
		hist.Open = stat.Open
		hist.Low = stat.Low
		hist.Volume = stat.Volume
		hists = append(hists, hist)
	}
	ShowResponse(w, AR{"prices": hists}, "", http.StatusOK)

}

func main() {

	conf := &config.Config{}
	config.GetConf(conf)

	dsn := conf.BirgaAddr
	var err error
	DB, err = bdaccess.GetDBClient()
	if err != nil {
		panic(err.Error())
	}
	defer DB.Close()

	fmt.Println(dsn)
	grcpConn, err := grpc.Dial(
		dsn,
		grpc.WithInsecure(),
	)
	if err != nil {
		fmt.Println("cant connect to grpc")
	}
	defer grcpConn.Close()

	brokerClient = info.NewExchangeClient(grcpConn)

	myBrokerID = &info.BrokerID{
		ID: 1,
	}

	go StartListen(brokerClient, myBrokerID)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/status", ShowStatReq)
	r.HandleFunc("/api/v1/deal", PushDeal)
	r.HandleFunc("/api/v1/cancel", CancelDeal)
	r.HandleFunc("/api/v1/history", ShowHistory).Queries("ticker", "{ticker}")

	fmt.Println("starting server at :5050")
	err = http.ListenAndServe(":5050", r)
	if err != nil {
		fmt.Printf("We have error: %s\n", err.Error())
	}
}
