package main

import (
	"fmt"
)

func main() {
	test := make(chan string)

	go func() {
		for i := 0; i < 5; i++ {
			test <- "hey"
		}
	}()

	select {
	case <-test:
		fmt.Println("here")
	}
}
