package main

import (
	"context"
	"errors"
	"github.com/tsmweb/chasam/app/fsstream"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/pkg/phash"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()
	log.Println("[>] Start")

	ctx, cancelFun := context.WithCancel(context.Background())
	roots := []string{
		"/home/martins/Desenvolvimento/SPTC/files/test",
	}

	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		cancelFun()
	}()

	fsstream.NewFileSearchStream(ctx, roots).
		OnError(func(err error) {
			log.Printf("[!] %v\n", err.Error())
		}).
		OnPipe(pipeKeyword).
		OnPipe(pipeSHA1).
		OnPipe(pipeED2K).
		OnPipe(pipeFilter).
		OnPipe(pipeAHash).
		OnPipe(pipeDHash).
		OnPipe(pipePHash).
		OnMatch(func(m media.Media) {
			log.Printf("[v] %20s - MATCH\n", m.Name)
		})

	elapsed := time.Since(start)
	log.Printf("[>] Stop - %s\n", elapsed)

	time.Sleep(time.Second)
	panic(errors.New("error"))
}

func pipeFilter(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	if m.Type == "image" {
		return fsstream.Next, nil
	}
	return fsstream.Skip, nil
}

func pipeKeyword(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	log.Printf("[*] %20s - KEYWORD\n", m.Name)
	return fsstream.Next, nil
}

func pipeSHA1(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	if err := m.GenSHA1(); err != nil {
		return fsstream.Skip, err
	}
	log.Printf("[*] %20s - SHA1[ %s ]\n", m.Name, m.SHA1)
	return fsstream.Next, nil
}

func pipeED2K(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	if err := m.GenED2K(); err != nil {
		return fsstream.Skip, err
	}
	log.Printf("[*] %20s - ED2K[ %s ]\n", m.Name, m.ED2K)
	return fsstream.Next, nil
}

func pipeAHash(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	if err := m.GenAHash(); err != nil {
		return fsstream.Skip, err
	}
	for _, h := range m.AHash {
		log.Printf("[*] %20s - A-HASH[ %s ]\n", m.Name, phash.FormatToHex(h))
	}

	return fsstream.Next, nil
}

func pipeDHash(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	if err := m.GenDHash(); err != nil {
		return fsstream.Skip, err
	}
	for _, h := range m.DHash {
		log.Printf("[*] %20s - D-HASH[ %s ]\n", m.Name, phash.FormatToHex(h))
	}

	return fsstream.Next, nil
}

func pipePHash(ctx context.Context, m media.Media) (fsstream.ResultType, error) {
	if err := m.GenPHash(); err != nil {
		return fsstream.Skip, err
	}
	for _, h := range m.PHash {
		log.Printf("[*] %20s - P-HASH[ %s ]\n", m.Name, phash.FormatToHex(h))
	}

	return fsstream.Match, nil
}
