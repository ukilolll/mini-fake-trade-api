package test

import "fmt"

func Test0() {
	var main []int
	temp := []int{1, 2, 4, 6, 3, 324, 343, 43}
	main = temp
	main = temp
	main = temp
	main = temp
	fmt.Println(main)
}