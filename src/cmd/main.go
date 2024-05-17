package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello World")

	for {
		t := time.Now()
		fmt.Println(t.Format("2006-01-02T15:04:05Z07:00"))
		time.Sleep(10 * time.Second)
	}
}
