package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type ClientRequest struct {
	Body   []byte
	Query  string
	Method string
}

func MakeRequest(brokerURL string, token string, clientReq ClientRequest) ([]byte, error) {

	var req *http.Request
	var err error
	client := &http.Client{Timeout: time.Second}

	if clientReq.Method == http.MethodPost {
		data, err := json.Marshal(clientReq.Body)
		if err != nil {
			return nil, fmt.Errorf("Bad Request")
		}
		reqBody := bytes.NewReader(data)
		req, err = http.NewRequest(clientReq.Method, brokerURL, reqBody)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Auth-Token", token)

	} else {

		req, err = http.NewRequest(clientReq.Method, brokerURL+"?"+clientReq.Query, nil)
		req.Header.Add("X-Auth-Token", token)
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

	switch response.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("Bad AccessToken")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("Internal Server Error")
	case http.StatusBadRequest:
		return nil, fmt.Errorf("Bad Request")
	}
	return body, nil

}
