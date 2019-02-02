package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Fruit struct {
	Name  string
	Color string
}

type Boof struct {
	Hehe  string
	Moing string
	Sproing []Fruit
}

func main() {
	bytearray, _ := ioutil.ReadFile("1.json")
	tprint(bytearray)
	fmt.Printf("%s\n", string(bytearray))

	// var jvar []interface{}
	jvar := make([]map[string]interface{},0)
	// var jvar []Fruit

	json.Unmarshal(bytearray, &jvar)
	cprint(jvar)
	cprint(jvar[1])
}

func tprint(v interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf("%T", v))
}

func cprint(v interface{}) {
	fmt.Printf("%#v\n", v)
}
