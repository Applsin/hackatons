package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hakaton-2018-2-2-mephi/broker/bdaccess"
	"hakaton-2018-2-2-mephi/broker/info"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
)

type CustomDeal struct {
	MyDeal CustomDealBody `json:"deal"`
}

type CustomDealBody struct {
	Ticker string  `json:"ticker"`
	Type   string  `json:"type"`
	Amount int     `json:"amount"`
	Price  float32 `json:"price"`
}

type DealAns struct {
	Body DealAnsBody `json:"body"`
}

type DealAnsBody struct {
	ID int64 `json:"id"`
}

type CustomCancel struct {
	Body CustomCancelBody `json:"body"`
}

type CustomCancelBody struct {
	ID     int64 `json:"id`
	Status bool  `json:"status"`
}

func insertOHLCV(ohlcv *info.OHLCV) {
	newOHLCV := &bdaccess.Stat{
		ID:       ohlcv.GetID(),
		Time:     ohlcv.GetTime(),
		Interval: ohlcv.GetInterval(),
		Open:     ohlcv.GetOpen(),
		High:     ohlcv.GetHigh(),
		Low:      ohlcv.GetLow(),
		Close:    ohlcv.GetClose(),
		Volume:   ohlcv.GetVolume(),
		Ticker:   ohlcv.GetTicker(),
	}

	DB.CreateStat(newOHLCV)
}

func PushDeal(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PushDeal")
	client := DB.SelectClientByLogin(r.Header.Get("X-Login"))
	if client.ID == 0 {
		ShowResponse(w, AR{}, "unauthorized", 403)
		fmt.Println(client)
		return
	}
	fmt.Println("client is ok")
	body, _ := ioutil.ReadAll(r.Body)
	deal := &CustomDeal{}
	err := json.Unmarshal(body, &deal)
	if err != nil {
		fmt.Println("ERROR PushDeal: ", err.Error())
		return
	}
	ctx := context.Background()
	//ctx = context.WithValue(ctx, "TYPE", deal.MyDeal.Type)
	md := metadata.Pairs(
		"TYPE", deal.MyDeal.Type,
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	//clientId, _ := strconv.Atoi(r.Header.Get("X-Token"))

	ras, err := brokerClient.Create(ctx, &info.Deal{
		ID:       currID,
		Ticker:   deal.MyDeal.Ticker,
		BrokerID: int32(myBrokerID.GetID()),
		ClientID: int32(DB.SelectClientIdByLogin(r.Header.Get("X-Login"))),
		Amount:   int32(deal.MyDeal.Amount),
		Time:     int32(time.Now().Unix()),
		Price:    deal.MyDeal.Price,
	})
	if err != nil {
		fmt.Println("FUCK UP " + err.Error())
	}
	fmt.Println(deal)
	// Сходить в базу и создать request, position
	DB.CreatePositions(&bdaccess.Position{
		ClientID: int32(DB.SelectClientIdByLogin(r.Header.Get("X-Login"))),
		Ticker:   deal.MyDeal.Ticker,
		Vol:      int32(deal.MyDeal.Amount),
	})
	dealAns := &DealAns{
		Body: DealAnsBody{
			ID: ras.ID,
		},
	}
	// need to set Bought properly
	DB.CreateRequest(&bdaccess.Request{
		ClientId: int64(DB.SelectClientIdByLogin(r.Header.Get("X-Login"))),
		Ticker:   deal.MyDeal.Ticker,
		Vol:      deal.MyDeal.Amount,
		Price:    deal.MyDeal.Price,
		Bought:   1,
	})
	currID++

	res, err := json.Marshal(dealAns)
	fmt.Println(res)
	w.WriteHeader(http.StatusFound)
	w.Write(res)

}

func CancelDeal(w http.ResponseWriter, r *http.Request) {

	/*
	   id1 := SelectIdByLogin(r.Header.Get("X-Login"))
	   if id1 == -1 {
	       ShowResponse(w, AR{}, "unauthorized", 403)
	       return
	   }
	*/

	fmt.Println("Cancel Deal")
	body, _ := ioutil.ReadAll(r.Body)
	id := &DealAnsBody{}
	err := json.Unmarshal(body, id)
	if err != nil {
		fmt.Println("ERROR CancelDeal: ", err.Error())
	}

	ctx := context.Background()

	result, err := brokerClient.Cancel(ctx, &info.DealID{
		BrokerID: myBrokerID.GetID(),
		ID:       id.ID,
	})
	if err != nil {
		fmt.Println("ERROR CancelDeal: ", err.Error())
	}

	myResult := &CustomCancel{
		Body: CustomCancelBody{
			ID:     id.ID,
			Status: result.GetSuccess(),
		},
	}

	DB.DeleteReqById(id.ID)
	DB.DeletePosById(id.ID)

	ans, err := json.Marshal(myResult)
	fmt.Println("Answer: ", string(ans))
	w.Write(ans)
}

func StartListen(brokerClient info.ExchangeClient, id *info.BrokerID) {

	ctx := context.Background()
	client, _ := brokerClient.Statistic(ctx, myBrokerID)

	go func() {
		for {
			inOHLCV, err := client.Recv()
			if err == io.EOF {
				fmt.Println("\tstream closed")
				return
			} else if err != nil {
				fmt.Println("\terror happed", err)
				return
			}
			insertOHLCV(inOHLCV)
		}

	}()

	results, err := brokerClient.Results(ctx, myBrokerID)
	if err != nil {
		fmt.Printf("results error: %s\n", err.Error())
	}

	go func() {
		for {
			fmt.Println("RESULTS:")
			result, err := results.Recv()
			if err == io.EOF {
				fmt.Println("\tstream closed")
				return
			} else if err != nil {
				fmt.Println("\terror happed", err)
				return
			}
			// Записать в базу
			fmt.Println(result)
		}

	}()
}
