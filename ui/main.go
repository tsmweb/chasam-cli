package main

import (
	"context"
	"errors"
	"github.com/tsmweb/chasam/app"
	"github.com/tsmweb/chasam/app/media"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()
	log.Println("[>] Start")

	roots := []string{
		"/home/martins/Desenvolvimento/SPTC/files",
	}

	fstream := app.NewFileStream(roots)

	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		fstream.Stop()
	}()

	fstream.
		OnError(func(err error) {
			log.Printf("[!] %v\n", err.Error())
		}).
		OnPipe(searchKeyword).
		OnPipe(searchSHA1).
		OnPipe(searchED2K).
		OnPipe(searchAHash).
		OnPipe(searchDHash).
		OnPipe(searchPHash).
		OnMatch(func(ctx context.Context, m media.Media) {
			log.Printf("\t\t[OK] MATCH - %s\n", m.Name)
		})
	
	elapsed := time.Since(start)
	log.Printf("\n[>] Stop - %s\n", elapsed)

	time.Sleep(time.Second)
	panic(errors.New("error"))
}

func searchKeyword(ctx context.Context, m media.Media) bool {
	//log.Printf("[>] KEYWORD - %s\n", m.Name)
	return false
}

func searchSHA1(ctx context.Context, m media.Media) bool {
	//log.Printf("[>] SHA1 - %s\n", m.Name)
	return false
}

func searchED2K(ctx context.Context, m media.Media) bool {
	//log.Printf("[>] ED2K - %s\n", m.Name)
	return false
}

func searchAHash(ctx context.Context, m media.Media) bool {
	//log.Printf("[>] A-HASH - %s\n", m.Name)
	return false
}

func searchDHash(ctx context.Context, m media.Media) bool {
	//log.Printf("[>] D-HASH - %s\n", m.Name)
	return false
}

func searchPHash(ctx context.Context, m media.Media) bool {
	//log.Printf("[>] P-HASH - %s\n", m.Name)
	return true
}
