package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ClientRequest struct {
	Body   string
	Query  string
	Method string
}

type Body struct {
	Body interface{} `json:"body"`
}

func MakeRequest(brokerURL string, login string, clientReq ClientRequest) ([]byte, error) {
	var req *http.Request
	var err error
	client := &http.Client{Timeout: time.Second}
	fmt.Println("Make Request")
	fmt.Println(brokerURL + clientReq.Query)
	if clientReq.Method == http.MethodPost {
		reqBody := bytes.NewBufferString(clientReq.Body)
		req, err = http.NewRequest(http.MethodPost, brokerURL+clientReq.Query, reqBody)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Length", strconv.Itoa(len(clientReq.Body)))
		req.Header.Add("X-Login", login)

	} else {

		req, err = http.NewRequest(http.MethodGet, brokerURL+clientReq.Query, nil)
		req.Header.Add("X-Login", login)
	}

	response, err := client.Do(req)

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, fmt.Errorf("timeout")
		}
		return nil, fmt.Errorf("unknown error %s", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	b := Body{}
	fmt.Println(string(body))
	err = json.Unmarshal([]byte(body), &b)
	if err != nil {
		return nil, fmt.Errorf("can't unpack json : " + err.Error())
	}
	res, _ := json.Marshal(b.Body)
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("Bad AccessToken")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("Internal Server Error")
	case http.StatusBadRequest:
		return nil, fmt.Errorf("Bad Request")
	}
	return res, nil
}

func GetHistory(ticker string) ClientRequest {
	return ClientRequest{
		Query:  "/api/v1/history?ticker=" + ticker,
		Method: http.MethodGet,
	}
}

func GetStatus() ClientRequest {
	return ClientRequest{
		Query:  "/api/v1/status",
		Method: http.MethodGet,
	}
}

func SubmitDeal(deal string) ClientRequest {
	result := ClientRequest{
		Query:  "/api/v1/deal",
		Method: http.MethodPost,
	}
	result.Body = "{\"deal\":" + deal + "}"

	fmt.Println(result)
	return result
}

func CancelDeal(id string) ClientRequest {
	result := ClientRequest{
		Query:  "/api/v1/cancel",
		Method: http.MethodPost,
	}
	result.Body = "{\"id\":" + id + "}"
	return result
}
