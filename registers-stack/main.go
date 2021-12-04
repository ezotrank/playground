package main

import "fmt"

//go:noinline
func add(i, j int) int {
	return i + j
}

func main() {
	z := add(22, 33)
	fmt.Println(z)
}
