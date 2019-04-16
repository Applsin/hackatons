package bdaccess

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//USAGE:
//dbClient, err := GetDBClient()
//if err != nil {
//    panic(err)
//}
//defer dbClient.Close()
//
// База - postgres Можно поднять локально, тогда меняйте dsn

type Client struct {
	ID      int64
	Login   string `gorm:"size:255; not null"`
	Balance int    `gorm:"not null"`
}

type Position struct {
	ID       int64
	ClientID int32  `gorm:"not null"`
	Ticker   string `gorm:"size:300; not null"`
	Vol      int32  `gorm:"not null"`
}

type OrderHistory struct {
	ID       int64
	Time     int64   `gorm:"not null"`
	ClientID int64   `gorm:"not null"`
	Ticker   string  `gorm:"size:300; not null"`
	Vol      int     `gorm:"not null"`
	Price    float32 `gorm:"not null"`
	Bought   int     `gorm:"not null"`
}

type Request struct {
	ID       int64
	ClientId int64   `gorm:"not null"`
	Ticker   string  `gorm:"size:300; not null"`
	Vol      int     `gorm:"not null"`
	Price    float32 `gorm:"not null"`
	Bought   int     `gorm:"not null"`
}

type Stat struct {
	ID       int64
	Time     int32
	Interval int32
	Open     float32
	High     float32
	Low      float32
	Close    float32
	Volume   int64
	Ticker   string `gorm:"size:300"`
}

type DatabaseAccessor struct {
	db *gorm.DB
}

//Изменение базы
/*func main() {
	db, _ := gorm.Open("postgres", "host=rc1c-4vhudj8phumy2c8w.mdb.yandexcloud." +
		"net port=6432 user=alexandr dbname=go_hackaton password=sL93U~Tq sslmode=verify-full")
	db.Set("gorm:table_options", "").AutoMigrate(&Client{}, &Position{}, &OrderHistory{}, &Request{}, &Stat{})
}*/

func GetDBClient() (*DatabaseAccessor, error) {
	db, err := gorm.Open("postgres", "host=localhost user=alexandr dbname=go_hackaton password=sL93U~Tq")

	if err != nil {
		return nil, err
	}
	db.Set("gorm:table_options", "").AutoMigrate(&Client{}, &Position{}, &OrderHistory{}, &Request{}, &Stat{})
	return &DatabaseAccessor{db}, nil
}

func (db *DatabaseAccessor) Close() {
	db.db.Close()
}

func (db *DatabaseAccessor) SelectStatByTicker(ticker string) []Stat {
	var stats []Stat
	db.db.Table("stats").Find(&stats).Where("ticker = ?", ticker)
	return stats
}

func (db *DatabaseAccessor) SelectClientByLogin(login string) Client {
	var client Client
	db.db.Table("clients").First(&client).Where("login = ?", login)
	return client
}

func (db *DatabaseAccessor) SelectClientByID(id int64) Client {
	var client Client
	db.db.Table("clients").First(&client).Where("ID = ?", id)
	return client
}

func (db *DatabaseAccessor) SelectPositionsByClientId(client_id int64) []Position {
	var positions []Position
	db.db.Table("positions").Find(&positions).Where("client_id = ?", client_id)
	return positions
}

func (accessor *DatabaseAccessor) SelectClientIdByLogin(login string) int64 {
	var client Client
	accessor.db.Table("clients").First(&client).Where("login = ?", login)
	return client.ID
}

func (accessor *DatabaseAccessor) SelectRequestsByClientId(id int64) []Request {
	var reqs []Request
	accessor.db.Table("requests").Find(&reqs).Where("client_id = ?", id)
	return reqs
}

func (accessor *DatabaseAccessor) SelectStatsByTickerAndTime(ticker string, time int64) []Stat {
	var stats []Stat
	accessor.db.Table("stats").Find(&stats).Where("ticker = ? and time = ?", ticker, time)
	return stats
}

func (db *DatabaseAccessor) SelectBalanceByClientId(id int64) int {
	var client Client
	db.db.Table("clients").Find(&client).Where(id)
	return client.Balance
}

func (db *DatabaseAccessor) CreateStat(st *Stat) {
	db.db.Table("stats").Save(st)
}

func (db *DatabaseAccessor) CreateRequest(req *Request) {
	db.db.Table("requests").Save(req)
}

func (db *DatabaseAccessor) CreatePositions(pos *Position) {
	db.db.Table("positions").Save(pos)
}

func (db *DatabaseAccessor) DeleteReqById(id int64) {
	db.db.Table("requests").Delete(Request{}).Where("id = ?", id)
}

func (db *DatabaseAccessor) DeletePosById(id int64) {
	db.db.Table("positions").Delete(Position{}).Where("id = ?", id)
}
