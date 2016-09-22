package main

import (
	"log"
	"sync"

	"github.com/thomasf/electro/pkg/ut61"
	ut61usb "github.com/thomasf/electro/pkg/ut61/usb"
)

func main() {

	err := ut61usb.CheckSuspend()
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan []byte, 0)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := ut61usb.ReadAll("0000", ch)
		close(ch)
		if err != nil {
			log.Fatal(err)
		}
	}()

loop:
	for data := range ch {
		msg, err := ut61.Parse(data)
		if err != nil {
			log.Println(err)
			continue loop
		}
		log.Println(msg.String())

	}
}
