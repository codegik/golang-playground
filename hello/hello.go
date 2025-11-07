package main

import (
	"codegik.com/greetings"
	"fmt"
)

func main() {
	message := greetings.Hello("Gladys")
	fmt.Println(message)
}
