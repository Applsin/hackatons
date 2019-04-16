package main

import (
	"fmt"
	"strings"
)

type querry struct {
	typework int
	fields   []string
	group    string
}

func ParseString(input string) querry {
	var output querry
	strSlice := strings.Split(input, " ")
	switch strSlice[1] {
	case "stream":
		output.typework = -1
	case "stat":
		output.typework = 0
	case "online":
		output.typework = 1
	}
	for i := 2; i < len(strSlice); i++ {
		if strSlice[i] == "from" {
			break
		}
		output.fields = append(output.fields, strSlice[i])
	}
	output.group = strSlice[len(strSlice)-1]
	return output
}

func main() {
	test := ParseString("select stream id, time, title from amazon_meta where groop = Book")
	fmt.Println(test)
}
