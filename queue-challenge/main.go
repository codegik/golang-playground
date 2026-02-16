package main

import "fmt"

func main() {
	stations1 := []int{1, 2, 4, 5, 0}
	r1 := 1
	k1 := 2
	result1 := maxPower(stations1, r1, k1)
	fmt.Printf("Input: stations = %v, r = %d, k = %d\n", stations1, r1, k1)
	fmt.Printf("Output: %d\n\n", result1)

	stations2 := []int{4, 4, 4, 4}
	r2 := 0
	k2 := 3
	result2 := maxPower(stations2, r2, k2)
	fmt.Printf("Input: stations = %v, r = %d, k = %d\n", stations2, r2, k2)
	fmt.Printf("Output: %d\n\n", result2)
}
