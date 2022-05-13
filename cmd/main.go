package main

import (
	"context"
	"errors"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/common"
	"github.com/tsmweb/chasam/pkg/ebus"
	"log"
	"os"
	"time"
)

func main() {
	log.Println("[>] Start")
	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		cancelFunc()
	}()

	foundMediaCH := make(chan *media.Media)

	go printError(ctx)
	go printFoundMedia(ctx)
	go printMatch(ctx, foundMediaCH)

	a := media.NewSearch(ctx)
	a.Execute([]string{"/home/martins/Desenvolvimento/SPTC/files/original"}, foundMediaCH)

	log.Println("[>] Stop")
	time.Sleep(time.Second * 3)
	panic(errors.New("error"))
}

func printFoundMedia(ctx context.Context) {
	sub := ebus.Instance().Subscribe(common.MediaTopic)
	defer sub.Unsubscribe()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case event, ok := <-sub.Event:
			if !ok {
				break loop
			}
			go func(e ebus.DataEvent) {
				m := e.Data.(*media.Media)
				log.Printf("[>] %s - %s\n", m.Name, m.Type)
			}(event)
		}
	}
}

func printError(ctx context.Context) {
	sub := ebus.Instance().Subscribe(common.ErrorTopic)
	defer sub.Unsubscribe()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case event, ok := <-sub.Event:
			if !ok {
				break loop
			}
			go func(e ebus.DataEvent) {
				log.Printf("[!] %v\n", e.Data.(error))
			}(event)
		}
	}
}

func printMatch(ctx context.Context, matchCH <-chan *media.Media) {
loop:
	for {
		select {
		case <-ctx.Done():
			// Drain matchCH to allow existing goroutines to finish.
			for range matchCH {
				// Do nothing.
			}
			return
		case m, ok := <-matchCH:
			if !ok {
				break loop
			}
			log.Printf("[>] Match: %s\n", m.Name)
		}
	}
}
