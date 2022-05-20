package main

import (
	"context"
	"errors"
	"github.com/tsmweb/chasam/app/fstream"
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
		"/home/martins/Desenvolvimento/SPTC/files/benchmark",
	}

	go func() {
		os.Stdin.Read(make([]byte, 1)) // read a single byte
		cancelFun()
	}()

	fstream.NewFileSearchStream(ctx, roots).
		OnError(func(err error) {
			log.Printf("[!] %v\n", err.Error())
		}).
		OnEach(pipeFilter).
		//OnEach(pipeKeyword).
		OnEach(pipeSHA1).
		OnEach(pipeED2K).
		OnEach(pipeAHash).
		OnEach(pipeDHash).
		OnEach(pipePHash).
		OnMatch(func(m *media.Media) {
			hSha1, _ := m.SHA1()
			hEd2k, _ := m.ED2K()

			aHash, _ := m.AHash()
			aHashStr := ""
			if len(aHash) > 0 {
				aHashStr = phash.FormatToHex(aHash[0])
			}

			dHash, _ := m.DHash()
			dHashStr := ""
			if len(dHash) > 0 {
				dHashStr = phash.FormatToHex(dHash[0])
			}

			pHash, _ := m.PHash()
			pHashStr := ""
			if len(pHash) > 0 {
				pHashStr = phash.FormatToHex(pHash[0])
			}

			log.Printf("[v] %s \n\tSHA1 [ %s ] \n\tED2K [ %s ] \n\tA-HASH [ %s ] "+
				"\n\tD-HASH [ %s ]\n\tP-HASH [ %s ]\n",
				m.Name(), hSha1, hEd2k, aHashStr, dHashStr, pHashStr)
		})

	elapsed := time.Since(start)
	log.Printf("[>] Stop - %s\n", elapsed)

	time.Sleep(time.Second)
	panic(errors.New("error"))
}

func pipeFilter(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if m.Type() == "image" {
		return fstream.Next, nil
	}
	return fstream.Skip, nil
}

func pipeKeyword(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	//log.Printf("[*] %20s - KEYWORD\n", m.Name)
	return fstream.Next, nil
}

func pipeSHA1(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if _, err := m.SHA1(); err != nil {
		return fstream.Skip, err
	}
	return fstream.Next, nil
}

func pipeED2K(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if _, err := m.ED2K(); err != nil {
		return fstream.Skip, err
	}

	if m.Type() == "video" {
		return fstream.Match, nil
	}

	return fstream.Next, nil
}

func pipeAHash(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if _, err := m.AHash(); err != nil {
		return fstream.Skip, err
	}

	return fstream.Next, nil
}

func pipeDHash(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if _, err := m.DHash(); err != nil {
		return fstream.Skip, err
	}

	return fstream.Next, nil
}

func pipePHash(_ context.Context, m *media.Media) (fstream.ResultType, error) {
	if _, err := m.PHash(); err != nil {
		return fstream.Skip, err
	}

	return fstream.Match, nil
}
