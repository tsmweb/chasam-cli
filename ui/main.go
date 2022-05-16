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
		"/home/martins/Desenvolvimento/SPTC/files/test",
	}

	fstream := app.NewFileSearchStream(roots)

	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		fstream.Stop()
	}()

	fstream.
		OnError(func(err error) {
			log.Printf("[!] %v\n", err.Error())
		}).
		OnLog(func(l app.Log) {
			log.Printf("[i] %s - %s - %s\n", l.FileName, l.FileType, l.FilePath)
		}).
		OnPipe(searchKeyword).
		OnPipe(searchSHA1).
		OnPipe(searchED2K).
		OnPipe(searchAHash).
		OnPipe(searchDHash).
		OnPipe(searchPHash).
		OnFound(func(ctx context.Context, m media.Media) {
			log.Printf("\t\t[v] MATCH - %s\n", m.Name)
		})

	elapsed := time.Since(start)
	log.Printf("[>] Stop - %s\n", elapsed)

	time.Sleep(time.Second)
	panic(errors.New("error"))
}

func searchKeyword(ctx context.Context, m media.Media) (bool, error) {
	//log.Printf("[>] KEYWORD - %s\n", m.Name)
	return false, nil
}

func searchSHA1(ctx context.Context, m media.Media) (bool, error) {
	//log.Printf("[>] SHA1 - %s\n", m.Name)
	return false, nil
}

func searchED2K(ctx context.Context, m media.Media) (bool, error) {
	//log.Printf("[>] ED2K - %s\n", m.Name)
	return false, nil
}

func searchAHash(ctx context.Context, m media.Media) (bool, error) {
	//log.Printf("[>] A-HASH - %s\n", m.Name)
	return false, nil
}

func searchDHash(ctx context.Context, m media.Media) (bool, error) {
	//log.Printf("[>] D-HASH - %s\n", m.Name)
	return false, nil
}

func searchPHash(ctx context.Context, m media.Media) (bool, error) {
	//log.Printf("[>] P-HASH - %s\n", m.Name)
	return true, nil
}
