package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ear7h/json-select"
)

const data = `
{
	"greeting" : "Welcome to Good Burger, home of the Good Burger may I take your order",
	"menu" : [
		{ "name" : "Good Buger", "price" : 2 },
		{ "name" : "Good Shake", "price" : 1 }
	]
}
`

func main() {
	var v json_select.Selecter

	err := json.Unmarshal([]byte(data), &v.V)
	if err != nil {
		log.Fatal(err)
	}

	names, err := v.SelectSlice("menu", []int{}, "name")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", names)
}
