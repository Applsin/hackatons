package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Record struct {
	Id        int
	ASIN      string
	Title     string
	Group     string
	Salesrank string
}

func Read(out chan Record, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(os.Args[1])
	defer file.Close()
	if err != nil {
		panic("Cannot read")
	}
	reader := bufio.NewReader(file)
LOOP_OUT:
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break LOOP_OUT
			}
			panic("AAAAAAAAA")
		}
		parts := strings.Split(line, ":")
		if TrimString(parts[0]) != "Id" {
			continue LOOP_OUT
		}
		currentRecord := Record{}
		idString := TrimString(parts[1])
		result, err := strconv.Atoi(idString)
		if err != nil {
			panic("Id is not int")
		}
		currentRecord.Id = result
	LOOP_IN:
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				panic("AAAAAAAAA")
			}
			parts := strings.Split(line, ":")
			header := TrimString(parts[0])
			switch header {
			case "ASIN":
				currentRecord.ASIN = TrimString(parts[1])
			case "title":
				for i := 1; i < len(parts); i++ {
					currentRecord.Title += parts[i]
					if i != len(parts)-1 {
						currentRecord.Title += ":"
					}
				}
				currentRecord.Title = TrimString(currentRecord.Title)
			case "group":
				currentRecord.Group = TrimString(parts[1])
			case "salesrank":
				currentRecord.Salesrank = TrimString(parts[1])
			default:
				break LOOP_IN
			}
		}
		//fmt.Printf("id=%d, asin=%s, title=%s, group=%s, sr=%s\n", currentRecord.Id, currentRecord.ASIN,
		//	currentRecord.Title, currentRecord.Group, currentRecord.Salesrank)
		//здесь посылаем
		out <- currentRecord
		time.Sleep(time.Second)
	}
	close(out)
}

func TrimString(str string) string {
	return strings.Trim(str, " \r\n")
}

func main() { //ARGS: 1-file path; 2-ip; 3-first port; 4-number of workers; 5-salt
	wg := &sync.WaitGroup{}
	channel := make(chan Record)
	fp, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic("bad comand line arguments")
	}
	num, err := strconv.Atoi(os.Args[4])
	if err != nil {
		panic("bad comand line arguments")
	}
	httpIp := "http://" + os.Args[2] + ":"
	meta := &Meta{
		HttpIp:     httpIp,
		WritersNum: num,
		FirstPort:  fp,
		Salt:       "",
	}
	wg.Add(2)
	go Read(channel, wg)
	go Send(meta, channel, wg)
	wg.Wait()
}
