package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	BirgaAddr   string   `json:"birga_addres"`
	MaxConnects int      `json:"max_connects"`
	Tickers     []string `json:"tickers"`
}

func GetConf(config *Config) {
	f, err := os.Open("config/config.json")
	if err != nil {
		panic(err.Error())
	}
	sc := bufio.NewReader(f)
	line, _, _ := sc.ReadLine()
	err = json.Unmarshal(line, &config)
	fmt.Println(config)
}
