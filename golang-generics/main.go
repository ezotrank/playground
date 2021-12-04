package main

import (
	"fmt"

	"constraints"
)

func PrintAnything[T any](arg T) {
	fmt.Println(arg)
}

type stringer interface {
	String() string
}

type myString string

func (s myString) String() string {
	return string(s)
}

func PrintStringer[T stringer](arg T) {
	fmt.Println(arg.String())
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Equal[T comparable](a, b T) bool {
	return a == b
}

type MyComparable interface {
	~int | ~int64
}

func EqualMyComparable[T MyComparable](a, b T) bool {
	return a == b
}

type Bunch[T any] []T

func (b Bunch[T]) Add(item T) Bunch[T] {
	return append(b, item)
}

type Queue[T any] []T

func (q *Queue[T]) pub(v T) {
	*q = append(*q, v)
}

func (q *Queue[T]) sub() (T, bool) {
	if len(*q) == 0 {
		var v T
		return v, false
	}
	v := (*q)[0]
	*q = (*q)[1:]
	return v, true
}

type myNumber int

type Number interface {
	int8 | ~int
}

func sum[T Number](a, b T) T {
	return a + b
}

func main() {
	PrintAnything("hello world!")
	PrintAnything(123)
	println("---")

	PrintStringer(myString("A"))
	println("---")

	fmt.Println(Max(1, 2))
	fmt.Println(Max(11.11, 2.11))
	println("---")

	fmt.Println(Equal(1, 2))
	fmt.Println(Equal(1.11, 2.22))
	println("---")

	fmt.Println(EqualMyComparable(1, 2))
	println("---")

	fmt.Println(Bunch[int32]{1, 2, 2})
	fmt.Println(Bunch[string]{"a", "b", "c"})
	println("---")

	bunch := Bunch[int]{1, 2, 3}
	fmt.Println(bunch.Add(4))
	println("---")

	q := Queue[int]{}
	q.pub(1)
	if val, ok := q.sub(); ok {
		fmt.Println(val + 1)
	}
	println("---")

	fmt.Println(sum[int](1, 2))
	fmt.Println(sum[int8](1, 2))
	fmt.Println(sum[myNumber](1, 2))
}
