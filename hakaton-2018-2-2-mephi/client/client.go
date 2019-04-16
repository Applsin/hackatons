package main

import (
	"fmt"
	// "io/ioutil"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	API  string
	Tmpl *template.Template
}

type Price struct {
	Time   int32   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"vol"`
}

type ShowMyOrders struct {
	Balance    int        `json:"balance"`
	Positions  []Position `json:"positions"`
	OpenOrders []Order    `json:"open_orders"`
}

type ShowPrices struct {
	Prices []Price `json:"prices"`
	Ticker string  `json:"ticker"`
}

type Order struct {
	ID     int     `json:"id"`
	Ticker string  `json:"ticker"`
	Volume float64 `json:"vol"`
	Type   string  `json:"type"`
	Status int     `json:"status"`
}

type Position struct {
	Ticker string  `json:"ticker"`
	Volume float64 `json:"vol"`
	Type   string  `json:"type"`
}

func (h *Handler) CheckAuth(r *http.Request) (string, error) {
	login, err := r.Cookie("login")
	loggedIn := (err != http.ErrNoCookie)

	if !loggedIn {
		return "", fmt.Errorf("unauthorized")
	}

	return login.Value, nil
}

func (h *Handler) handleMain(w http.ResponseWriter, r *http.Request) {
	_, err := h.CheckAuth(r)
	if err != nil {
		fmt.Fprintln(w, `<a href="/login">login</a>`)
		fmt.Fprintln(w, "You need to login")
		return
	}

	err = h.Tmpl.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		panic(err)
	}
}

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(10 * time.Hour)
	cookie := http.Cookie{
		Name:    "login",
		Value:   "rvasily",
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) logoutPage(w http.ResponseWriter, r *http.Request) {
	login, err := r.Cookie("login")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	login.Expires = time.Now().AddDate(0, 0, -1)
	login.Value = ""

	http.SetCookie(w, login)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) handleMyPositions(w http.ResponseWriter, r *http.Request) {
	login, err := h.CheckAuth(r)
	if err != nil {
		fmt.Fprintln(w, `<a href="/login">login</a>`)
		fmt.Fprintln(w, "You need to login")
		return
	}

	fmt.Println("handleMyPos")
	var orders ShowMyOrders
	res, err := MakeRequest(h.API, login, GetStatus())
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(res, &orders)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("can't unpack json" + err.Error()))
		return
	}
	fmt.Println("orders ", orders)

	err = h.Tmpl.ExecuteTemplate(w, "my-positions.html", orders)
	if err != nil {
		panic(err)
	}
}

func (h *Handler) handleShowPrices(w http.ResponseWriter, r *http.Request) {
	login, err := h.CheckAuth(r)
	if err != nil {
		fmt.Fprintln(w, `<a href="/login">login</a>`)
		fmt.Fprintln(w, "You need to login")
		return
	}
	ticker := r.URL.Query().Get("ticker")

	var prices ShowPrices
	res, err := MakeRequest(h.API, login, GetHistory(ticker))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	err = json.Unmarshal(res, &prices)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("can't unpack json : " + err.Error()))
		return
	}

	err = h.Tmpl.ExecuteTemplate(w, "prices.html", prices)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("internal server error"))
		return
	}
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	login, err := h.CheckAuth(r)
	if err != nil {
		fmt.Fprintln(w, `<a href="/login">login</a>`)
		fmt.Fprintln(w, "You need to login")
		return
	}
	ticker := r.FormValue("ticker")
	typ := r.FormValue("type")
	amount := r.FormValue("amount")
	price := r.FormValue("price")

	d := fmt.Sprintf(`{"ticker": "%s", "type": "%s", "amount": %s, "price": %s}`, ticker, typ, amount, price)

	fmt.Println(d)
	_, err = MakeRequest(h.API, login, SubmitDeal(d))

	http.Redirect(w, r, "/my-positions", http.StatusFound)
}

func (h *Handler) handleCancelDeal(w http.ResponseWriter, r *http.Request) {
	login, err := h.CheckAuth(r)
	if err != nil {
		fmt.Fprintln(w, `<a href="/login">login</a>`)
		fmt.Fprintln(w, "You need to login")
		return
	}
	_, err = MakeRequest(h.API, login, CancelDeal(r.FormValue("delete")))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, "/my-positions", http.StatusFound)
}

func main() {

	h := &Handler{
		"http://localhost:5050",
		template.Must(template.ParseGlob("./tmpl/*")),
	}

	r := mux.NewRouter()
	r.HandleFunc("/login", h.loginPage).Methods("GET")
	r.HandleFunc("/logout", h.logoutPage).Methods("GET")
	r.HandleFunc("/", h.handleMain).Methods("GET")
	r.HandleFunc("/", h.handleCreate).Methods("POST")
	r.HandleFunc("/my-positions", h.handleMyPositions).Methods("GET")
	r.HandleFunc("/my-positions", h.handleCancelDeal).Methods("POST")
	r.HandleFunc("/prices", h.handleShowPrices).Methods("GET")

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", r)
}
