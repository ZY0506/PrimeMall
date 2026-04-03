package main

import "fmt"

func test01(s []int) {
	s = append(s, 1)
	fmt.Println(s)
	return
}

func main() {
	var s = []int{1, 1, 1}
	fmt.Println(len(s))
	fmt.Println(cap(s))
	test01(s)
	fmt.Println(s[3])
}
