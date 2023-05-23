package main

import "fmt"

func main() {
	tt := make(map[string][]string)
	tt["22"] = append(tt["22"], "11")
	fmt.Println(tt)
}
