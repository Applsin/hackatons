package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"html/template"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type History struct {
	ID            string
	Query         string
	Work_time     time.Duration
	Results_Count int
	Active        string
	Operations    []string
}

type Result struct {
	ID        int
	Time      int
	Title     string
	Group     string
	ASIN      string
	Salesrank string
}

type historyData struct {
	ID     string
	Mode   string
	Fields []string
	Group  string
	Time   time.Time
}

type QueryData struct {
	Query      string
	Active     bool
	Operations string
	Id         int
	Time       int
	ASIN       string
	Title      string
	Group      string
	Salesrank  string
}

type Query struct {
	IDTask string
	Data   QueryData
}

var (
	tasks   map[string]historyData
	results map[string][]Result
	workers []net.Conn
)

const (
	RESULT = (
		`<html>
		<body>
			
			<h1>results</h1>
			<table>
			<th>ID</th>
			<th>Time</th>
			<th>Group</th>
			<th>Title</th>
			<th>ASIN</th>
			<th>Salesrank</th>

			{{range .Results}}
				<tr>
				<td>{{.ID}}</td>
				<td>{{.Time}}</td>
				<td>{{.Group}}</td>
				<td>{{.Title}}</td>
				<td>{{.ASIN}}</td>
				<td>{{.Salesrank}}</td>
				</tr>
			{{end}}

			</table>
		</body>
		</html>
	`)

	HISTORY = (`
		<html>
		<html>
		<body>
			<h1>history</h1>
			<table>
			{{range .History}}
				<tr>
					<td>{{.ID}}</td>
					<td>{{.Query}}</td>
					<td>{{.Work_time}}</td>
					<td>{{.Results_Count}}</td>
					<td>{{.Active}}</td>
					<td>
					{{range .Operations}}
						<span>{{.}}</span>
					{{end}}
					</td>
				</tr>
			{{end}}
			</table>
		</body>
		</html>
		</html>
	`)
)

func handleMainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
	<html>
	<div style="display: flex; align-items:  center; justify-content: center">
		<form id="form" action="/push">
			<input id="input" type="text">
			<input type="submit">
		</form>
	</div>
	<script>

		var form = document.getElementById("form");
		var input = document.getElementById("input");
		console.log(form, input.value);
		form.onsubmit = function(event) {
			event.preventDefault();
			window.location.href += "/push?input=" + input.value;
		};

	</script>
	</html>
	`))
}

func ReadResponse(conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		text := scanner.Text()
		//fmt.Println("*****", text)

		handleResponse([]byte(text))
	}
}

func IsIdExist(r *Result) bool {
	return r.ID != 0
}

func IsTimeExist(r *Result) bool {
	return r.Time != 0
}

func IsTitleExist(r *Result) bool {
	return r.Title != ""
}

func IsGroupExist(r *Result) bool {
	return r.Group != ""
}

func IsASINExist(r *Result) bool {
	return r.ASIN != ""
}

func IsSalesrankExist(r *Result) bool {
	return r.Salesrank != ""
}

func main() {
	listner, err := net.Listen("tcp", "127.0.0.1:8090")

	if err != nil {
		fmt.Print("cant start tcp at 127.0.0.1:8090")
		os.Exit(1)
	}
	fmt.Println("starting listen tcp at :8090")

	go func() {
		for {
			conn, err := listner.Accept()
			fmt.Println("connect")
			if err != nil {
				fmt.Print("cant accept connection")
				os.Exit(1)
			}
			workers = append(workers, conn)
			go ReadResponse(conn)
		}
	}()

	http.HandleFunc("/", handleMainPage)

	resultTmpl := template.New(`result`)
	resultTmpl, _ = resultTmpl.Parse(RESULT)

	historyTmpl := template.New(`history`)
	historyTmpl, _ = historyTmpl.Parse(HISTORY)

	http.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		text := r.URL.Query().Get("input")
		fmt.Println(text)

		for _, conn := range workers {
			conn.Write(handleRequest(text))
		}

		http.Redirect(w, r, "/", 301)
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {

		var res []Result

		for _, val := range results {
			res = append(res, val...)
		}

		err := resultTmpl.Execute(w,
			struct {
				Results []Result
			}{
				res,
			})
		if err != nil {
			panic(err)
		}
	})

	// 	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
	// 		err := historyTmpl.Execute(w,
	// 			struct {
	// 				History []History
	// 			}{
	// 				tasks,
	// 			})
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	})

		fmt.Println("starting http at :8080")
		http.ListenAndServe(":8080", nil)
}

func writeHistory(query string) string {
	timestamp := time.Now()
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, uint64(timestamp.Unix()))
	hash := md5.Sum(b)
	var arr [16]byte
	copy(arr[:], hash[:16])
	mode, condition, fields := parseQuery(query)
	id := string(arr[:16])
	tasks[string(arr[:16])] = historyData{string(arr[:16]), mode, fields, condition, timestamp}
	return id
}

func parseQuery(query string) (mode, condition string, fields []string) {
	str := strings.Split(query, " ")
	mode = str[1]
	var fromIndex int
	for idx, v := range str {
		if v == "from" {
			fromIndex = idx
			break
		}
	}
	fields = str[2:fromIndex]
	for _, val := range fields {
		strings.Trim(val, ",")
	}
	condition = str[len(str)-1]
	return mode, condition, fields
}

//возвращает время работы в сек предполагается, что будет получать Time из historyData
func getWorkTime(startTime time.Time) time.Duration {
	return time.Since(startTime) / time.Second
}

//возвращает кол-во результатов по md5 id таска
func getResultCount(taskId string) int {
	return len(results[taskId])
}

func Pack(query *Query) ([]byte, error) {

	queryByte, err0 := json.Marshal(query)

	res := []byte{}
	idTask := new(bytes.Buffer)
	err1 := binary.Write(idTask, binary.LittleEndian, query.IDTask)

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
	fmt.Println(string(task))

	err0 := binary.Read(bytes.NewReader(task[:4]), binary.LittleEndian, &query.IDTask)
	fmt.Println(err0, query.IDTask)
	err1 := binary.Read(bytes.NewReader(task[4:8]), binary.LittleEndian, &idRecord)

	err2 := binary.Read(bytes.NewReader(task[8:12]), binary.LittleEndian, &lenData)

	err3 := json.Unmarshal(task[12:12+lenData], query)

	err4 := binary.Read(bytes.NewReader(task[12+lenData:16+lenData]), binary.LittleEndian, &src)

	if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || src != crc32.ChecksumIEEE(task[12:12+lenData]) {
		return fmt.Errorf("Err translation from bytes")
	}
	fmt.Println(query)
	return nil
}

func
handleResponse(res []byte) {
	query := Query{}
	Unpack(&query, res)
	taskId := query.IDTask
	queryData := query.Data
	queryParams := tasks[taskId]
	fields := queryParams.Fields
	result := getFieldsFromResponse(queryData, fields)
	results[string(taskId)] = append(results[string(taskId)], result)
}

func getFieldsFromResponse(response QueryData, fields []string) Result {
	var res Result
	for _, val := range fields {
		switch val {
		case "id":
			res.ID = response.Id
		case "time":
			res.Time = response.Time
		case "salesrank":
			res.Salesrank = response.Salesrank
		case "group":
			res.Group = response.Group
		case "title":
			res.Title = response.Title
		case "asin":
			res.ASIN = response.ASIN
		}
	}
	return res
}

func handleRequest(query string) []byte {
	fmt.Println("handleRequest called")
	taskId := writeHistory(query)
	fmt.Println("got task id")
	queryData := QueryData{Query: query}
	request, _ := Pack(&Query{taskId, queryData})

	return request
}
