package main

import (
	"context"
	"errors"
	"github.com/tsmweb/chasam/app"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/pkg/phash"
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
			//log.Printf("[i] %s - %s\n", l.FileName, l.FileType)
		}).
		OnFilter(func(m media.Media) bool {
			if m.Type == "image" {
				return true
			}
			return false
		}).
		OnSearch(searchKeyword).
		OnSearch(searchSHA1).
		OnSearch(searchED2K).
		OnSearch(searchAHash).
		OnSearch(searchDHash).
		OnSearch(searchPHash).
		OnMatch(func(m media.Media) {
			log.Printf("[v] %20s - MATCH\n", m.Name)
		})

	elapsed := time.Since(start)
	log.Printf("[>] Stop - %s\n", elapsed)

	time.Sleep(time.Second)
	panic(errors.New("error"))
}

func searchKeyword(ctx context.Context, m media.Media) (bool, error) {
	log.Printf("[*] %20s - KEYWORD\n", m.Name)
	return false, nil
}

func searchSHA1(ctx context.Context, m media.Media) (bool, error) {
	//if err := m.GenSHA1(); err != nil {
	//	return false, err
	//}
	log.Printf("[*] %20s - SHA1[ %s ]\n", m.Name, m.SHA1)
	return false, nil
}

func searchED2K(ctx context.Context, m media.Media) (bool, error) {
	//if err := m.GenED2K(); err != nil {
	//	return false, err
	//}
	log.Printf("[*] %20s - ED2K[ %s ]\n", m.Name, m.ED2K)
	return false, nil
}

func searchAHash(ctx context.Context, m media.Media) (bool, error) {
	//if err := m.GenAHash(); err != nil {
	//	return false, err
	//}
	for _, h := range m.AHash {
		log.Printf("[*] %20s - A-HASH[ %s ]\n", m.Name, phash.FormatToHex(h))
	}

	return false, nil
}

func searchDHash(ctx context.Context, m media.Media) (bool, error) {
	//if err := m.GenDHash(); err != nil {
	//	return false, err
	//}
	for _, h := range m.DHash {
		log.Printf("[*] %20s - D-HASH[ %s ]\n", m.Name, phash.FormatToHex(h))
	}

	return false, nil
}

func searchPHash(ctx context.Context, m media.Media) (bool, error) {
	//if err := m.GenPHash(); err != nil {
	//	return false, err
	//}
	for _, h := range m.PHash {
		log.Printf("[*] %20s - P-HASH[ %s ]\n", m.Name, phash.FormatToHex(h))
	}

	return true, nil
}
