package main

import (
	"encoding/json"
	"fmt"
)

func print(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
