package main

import (
	"fmt"
	"strings"
)

func main() {
	my_list := []string{"apple", "banana", "orange"}
	search := "banana"
	found := false
	for _, item := range my_list {
		if item == search {
			found = true
			break
		}
	}
	if found {
		fmt.Printf("The element %q is in the list.", search)
	} else {
		fmt.Printf("The element %q is not in the list.", search)
	}
	level := 10
	index := 111
	functionName := fmt.Sprintf("service-%d-%d", level, index)
	fmt.Println(functionName)
	fmt.Println(strings.Split(functionName, "-")[2])
}
