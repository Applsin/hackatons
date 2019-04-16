package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"time"
)

/* ToDo
сообщение, что таск завершен
*/

var config Config

// Структуры

type QueryDataWriter struct {
	Id        int    `json:"id"`
	Time      uint64 `json:"time"`
	ASIN      string `json:"asin"`
	Title     string `json:"title"`
	Group     string `json:"group"`
	Salesrank string `json:"salesrank"`
}

type QueryData struct {
	Query      string `json:"query"`
	Active     bool   `json:"active"`
	Operations string `json:"operations"`
	Id         int    `json:"id"`
	Time       uint64 `json:"time"`
	ASIN       string `json:"asin"`
	Title      string `json:"title"`
	Group      string `json:"group"`
	Salesrank  string `json:"salesrank"`
}

type Query struct {
	IDTask uint32
	Data   QueryData
}

func Pack(query *Query) ([]byte, error) {

	queryByte, err0 := json.Marshal(query)

	res := []byte{}
	idTask := new(bytes.Buffer)
	err1 := binary.Write(idTask, binary.LittleEndian, uint32(query.IDTask))

	idRecord := new(bytes.Buffer)
	err2 := binary.Write(idRecord, binary.LittleEndian, uint32(query.Data.Id))

	lenData := new(bytes.Buffer)
	err3 := binary.Write(lenData, binary.LittleEndian, uint32(len(queryByte)))

	Crc := new(bytes.Buffer)
	err4 := binary.Write(Crc, binary.LittleEndian, crc32.ChecksumIEEE(queryByte))
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return []byte{}, fmt.Errorf("Err translation in bytes")
	}

	res = append(res, idTask.Bytes()...)
	res = append(res, idRecord.Bytes()...)
	res = append(res, lenData.Bytes()...)
	res = append(res, queryByte...)
	res = append(res, Crc.Bytes()...)
	return res, nil
}

func Unpack(query *Query, task []byte) error {
	idRecord, lenData, src := uint32(0), uint32(0), uint32(0)
	//fmt.Println(string(task))

	err0 := binary.Read(bytes.NewReader(task[:4]), binary.LittleEndian, &query.IDTask)
	//fmt.Println(err0, query.IDTask)
	err1 := binary.Read(bytes.NewReader(task[4:8]), binary.LittleEndian, &idRecord)

	err2 := binary.Read(bytes.NewReader(task[8:12]), binary.LittleEndian, &lenData)

	err3 := json.Unmarshal(task[12:12+lenData], query)

	err4 := binary.Read(bytes.NewReader(task[12+lenData:16+lenData]), binary.LittleEndian, &src)

	if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || src != crc32.ChecksumIEEE(task[12:12+lenData]) {
		return fmt.Errorf("Err translation from bytes")
	}
	//fmt.Println(query)
	return nil
}

var ListTasks []Task
var WebAddr string
var filePath string

// ТАСК
type Task struct {
	NumbrStr uint32
	Number   uint32
	query    *Query
}

func Send(conn net.Conn, data []byte) {
	conn.Write(data)
	fmt.Println("SEND TO SERVER")
}

func СheckLine(line []byte, query *Query) bool {
	data := &QueryDataWriter{}
	json.Unmarshal(line, data)
	if data.Group == query.Data.Group {
		query.Data.Id = data.Id
		query.Data.Time = data.Time
		query.Data.ASIN = data.ASIN
		query.Data.Title = data.Title
		query.Data.Group = data.Group
		query.Data.Salesrank = data.Salesrank
		return true
	}
	return false
}

func Work(t chan Task, conn net.Conn) {
	timer := time.NewTimer(10 * time.Second)
	currId := 0
	f, _ := os.Open(config.FilePath)
	sc := bufio.NewReader(f)
	for {
		select {
		case task := <-t:
			fmt.Println("ADD TASK")
			ListTasks = append(ListTasks, task)
		case <-timer.C:
			timer.Reset(10 * time.Second)
			fmt.Println("SWITCH")
			if len(ListTasks) == 0 {
				fmt.Println("TASKS IS EMPTY")
			}
			if currId == len(ListTasks) {
				currId = 0
			} else {
				currId++
			}
			fmt.Printf("CURRID %d\n", currId)
			if len(ListTasks) > 0 {
				f, _ = os.Open(filePath)
				f.Seek(int64(ListTasks[currId].NumbrStr), 0)
				sc = bufio.NewReader(f)
				ListTasks[currId].NumbrStr++
			}
		default:
			if len(ListTasks) > 0 {
				fmt.Printf("CURRID %d\n", currId)
				line, _, err := sc.ReadLine()
				fmt.Printf("LINE : %s\n", line)
				if err != nil {
					if err == io.EOF {
						fmt.Println("END OF FILE")
						ListTasks = append(ListTasks[:currId], ListTasks[currId+1:]...)
						break
					}
				}
				if СheckLine(line, ListTasks[currId].query) {
					result, _ := Pack(ListTasks[currId].query)
					go Send(conn, result)
				}
			}

		}
	}
}

/*

func Work(t chan Task, conn net.Conn) {
	timer := time.NewTimer(10 * time.Second)
	currId := 0
	f, _ := os.Open(filePath)
	sc := bufio.NewReader(f)
	start := time.Now()
	counter := 0
	for {

		select {
		case task := <-t:
			fmt.Println("ADD TASK")
			ListTasks = append(ListTasks, task)
		case <-timer.C:
			timer.Reset(10 * time.Second)
			fmt.Println("SWITCH")
			if currId == len(ListTasks) {
				currId = 0
			} else {
				currId++
			}
			fmt.Printf("CURRID %d\n", currId)
			if len(ListTasks) > 0 {
				f, _ = os.Open(filePath)
				f.Seek(int64(ListTasks[currId].NumbrStr), 0)
				sc = bufio.NewReader(f)
				ListTasks[currId].NumbrStr++
			}
		default:
			if counter == 1000 {
				start = time.Now()
				counter = 0
				time.Sleep(time.Second - time.Since(start))
				fmt.Println("SLEEP HAPPENED")
			}
			if len(ListTasks) > 0 {
				fmt.Printf("CURRID %d\n", currId)
				line, _, err := sc.ReadLine()
				fmt.Printf("LINE : %s\n", line)
				if err != nil {
					if err == io.EOF {
						fmt.Println("END OF FILE")
						ListTasks = append(ListTasks[:currId], ListTasks[currId+1:]...)
						break
					}
				}
				if СheckLine(line, ListTasks[currId].query) {
					result, _ := Pack(ListTasks[currId].query)
					counter++
					go Send(conn, result)
				}
			}

		}
	}
}
*/

func AddTask(t chan Task, data []byte) {
	fmt.Println("FUNC ADD TASK")
	query := Query{}
	err := Unpack(&query, data)
	fmt.Println(query)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR ADD TASK")
		return
	}
	task := Task{
		0,
		query.IDTask,
		&query,
	}
	fmt.Println("GOING TO SEND")
	t <- task
	fmt.Println("SEND")
}

type Config struct {
	WebAddr   string `json:"web_addres"`
	FilePath  string `json:"file_path"`
	BufferLen int    `json:"buff_len"`
}

func getConf() {
	f, err := os.Open("config.json")
	if err != nil {
		panic(err.Error())
	}
	sc := bufio.NewReader(f)
	line, _, _ := sc.ReadLine()
	err = json.Unmarshal(line, &config)
	fmt.Println(config)
}

func main() {
	getConf()
	conn, err := net.Dial("tcp", config.WebAddr)
	if err != nil {
		fmt.Println("Cannot connect to server")
		return
	}
	fmt.Println("Connect was created")
	t := make(chan Task, config.BufferLen)
	go Work(t, conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := []byte(scanner.Text())
		fmt.Println(text)
		go AddTask(t, text)
	}
}
