package main

import "fmt"

func main() {
	th, f, err := NewTDHandler()
	if err != nil {
	}
	fmt.Println(th, err)
	f()
}
