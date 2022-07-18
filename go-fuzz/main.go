package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

func Reverse(s string) (string, error) {
	if !utf8.ValidString(s) {
		return s, errors.New("input is not valid UTF-8")
	}

	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return string(r), nil
}

func main() {
	rev, err := Reverse("Hello, world!")
	if err != nil {
		panic(err)
	}
	fmt.Printf("word: %s, reverse: %s\n", "Hello, world!", rev)
}
