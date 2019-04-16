package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"
	"time"
	"bytes"
	"strconv"
	// "encoding/json"
	"html/template"
)

type Handler struct {
	API string
	Tmpl *template.Template
}

type Price struct {
	Time string
	Open float64
	High float64
	Low float64
	Close float64
	Volume float64
}

type ShowMyOrders struct {
	Balance float64
	Positions []Position
	OpenOrders []Order
}

type ShowPrices struct {
	Prices []Price
	Ticker string
}

type Order struct {
	Id int
	Ticker string
	Volume float64
	Type string
	Status string
}

type Position struct {
	Ticker string
	Volume float64
	Type string
}

func requestBalance() {
	client := &http.Client{
		Timeout:   time.Second * 10,
	}

	data := `{"id": 42, "user": "rvasily"}`
	body := bytes.NewBufferString(data)

	url := "http://127.0.0.1:8080/api/v1/status"
	req, _ := http.NewRequest(http.MethodPost, url, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error happend", err)
		return
	}
	defer resp.Body.Close() // важный пункт!

	// respBody, err := ioutil.ReadAll(resp.Body)
}

func (h *Handler) handleOpenNew(w http.ResponseWriter, r *http.Request) {
	// res, _ := json.Marshal("main page")
	err := h.Tmpl.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		panic(err)
	}
}

func (h *Handler) handleMyPositions(w http.ResponseWriter, r *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "my-positions.html", ShowMyOrders{
			Balance: 10000000,
			Positions: []Position{
				Position{
					"SPFB.RTS",
					100,
					"BYE",
				},
				Position{
					"SPFB.RTS",
					100,
					"BYE",
				},
				Position{
					"SPFB.RTS",
					100,
					"BYE",
				},
			},
			OpenOrders: []Order{
				Order{9, "SPFB.RTS", 200, "SELL", "pending"},
				Order{9, "SPFB.RTS", 200, "SELL", "pending"},
				Order{9, "SPFB.RTS", 200, "SELL", "pending"},
				Order{9, "SPFB.RTS", 200, "SELL", "pending"},
			},
		})
	if err != nil {
		panic(err)
	}
}

func (h *Handler) handleShowPrices(w http.ResponseWriter, r *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "prices.html", ShowPrices{
		[]Price{
			{"12:00", 100, 200, 50, 1, 100},
			{"12:00", 100, 200, 50, 1, 100},
			{"12:00", 100, 200, 50, 1, 100},
			{"12:00", 100, 200, 50, 1, 100},
		},
		"SPFB.RTS",
	})
	if err != nil {
		panic(err)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleOpenNew(w, r)
	case "/open":
		h.handleOpenNew(w, r)
	case "/my-positions":
		h.handleMyPositions(w, r)
	case "/prices":
		h.handleShowPrices(w, r)
	}
}

func Client() *Handler {
	return &Handler{
		"http://localhost:8082/api/v1",
		template.Must(template.ParseGlob("./tmpl/*")),
	}
}

func main() {
	http.Handle("/", Client())

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
