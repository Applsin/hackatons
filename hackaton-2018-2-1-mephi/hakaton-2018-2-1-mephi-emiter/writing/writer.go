package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)


type WritingUnit struct {
	Id        int    `json:"id"`
	Time      int64  `json:"time"`
	ASIN      string `json:"asin"`
	Title     string `json:"title"`
	Group     string `json:"group"`
	Salesrank string `json:"salesrank"`
}

type Writer struct {
	One sync.Once
	Id  int
}

var mu = sync.Mutex{}

func runServer(addr string, i int) {
	mux := http.NewServeMux()
	writer := &Writer{Id: i}
	mux.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			writer.One.Do(func() {
				os.Mkdir("Worker"+strconv.Itoa(writer.Id), 0777)
				f, err := os.Create("Worker" + strconv.Itoa(writer.Id) + "/Worker" + strconv.Itoa(writer.Id) + ".txt")
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
			})
			body := &WritingUnit{}
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(data, body)
			body.Time = time.Now().Unix()
			mu.Lock()
			defer mu.Unlock()
			f, err := os.OpenFile("Worker"+strconv.Itoa(writer.Id)+"/Worker"+strconv.Itoa(writer.Id)+".txt", os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			result, err := json.Marshal(body)
			if err != nil {
				log.Fatal(err)
			}
			//log.Println(string(result[:len(result)]))
			_, err = io.WriteString(f, string(result)+"\n")
			//ioutil.WriteFile("Worker"+strconv.Itoa(writer.Id)+"/Worker"+strconv.Itoa(writer.Id)+".txt", result, 0777)
		})
	server := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	server.ListenAndServe()
}

func RunServ() {
	COUNT, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	SERV, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	for i := 0; i < COUNT-1; i++ {
		i := i
		go runServer("0.0.0.0:"+strconv.Itoa(SERV+i), i)
	}
	runServer("0.0.0.0:"+strconv.Itoa(SERV+COUNT-1), COUNT-1)
}

func main() {
	RunServ()
}
