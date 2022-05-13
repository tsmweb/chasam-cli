package media

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsmweb/chasam/common"
	"github.com/tsmweb/chasam/common/imageutil"
	"github.com/tsmweb/chasam/pkg/ebus"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Search interface {
	Execute(roots []string, foundCH chan<- *Media)
}

type search struct {
	ctx    context.Context
	semaCH chan struct{} // concurrency-limiting counting semaphore
}

func NewSearch(ctx context.Context) Search {
	return &search{
		ctx:    ctx,
		semaCH: make(chan struct{}, 20),
	}
}

func (s *search) Execute(roots []string, foundCH chan<- *Media) {
	// starts the processing pipeline.
	mediaCH := make(chan *Media)
	hashCH := make(chan *Media)
	phashCH := make(chan *Media)

	go s.searchByKeywork(hashCH, mediaCH, foundCH)
	go s.searchByHash(phashCH, hashCH, foundCH)
	go s.searchByPerceptualHash(phashCH, foundCH)

	var wg sync.WaitGroup

	for _, dir := range roots {
		wg.Add(1)
		go s.walkDir(dir, &wg, mediaCH)
	}

	wg.Wait()
	log.Println("[done] walkDir")
	close(mediaCH)
}

func (s *search) walkDir(dir string, wg *sync.WaitGroup, mediaCH chan<- *Media) {
	defer wg.Done()

	for _, entry := range s.dirents(dir) {
		if s.cancelled() {
			break
		}

		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go s.walkDir(subdir, wg, mediaCH)
		} else {
			m, err := New(filepath.Join(dir, entry.Name()))
			if err != nil {
				if !errors.Is(err, imageutil.ErrUnsupportedMediaType) {
					ebus.Instance().Publish(common.ErrorTopic, err)
				}
			} else {
				mediaCH <- m
			}
		}
	}
}

func (s *search) dirents(dir string) []os.FileInfo {
	select {
	case s.semaCH <- struct{}{}: // acquire token
	case <-s.ctx.Done():
		return nil // cancelled
	}

	defer func() { <-s.semaCH }() // release token

	f, err := os.Open(dir)
	if err != nil {
		ebus.Instance().Publish(common.ErrorTopic,
			fmt.Errorf("Search::dirents(%s) | Error: %v", dir, err))
		return nil
	}
	defer f.Close()

	entries, err := f.Readdir(0)
	if err != nil {
		ebus.Instance().Publish(common.ErrorTopic, err)
	}

	return entries
}

func (s *search) cancelled() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

func (s *search) searchByKeywork(hashCH chan<- *Media, mediaCH <-chan *Media, foundCH chan<- *Media) {
	defer func() {
		close(hashCH)
		log.Println("[done] searchByKeywork")
	}()

	for m := range mediaCH {
		if work("keyword", m) {
			foundCH <- m
		} else {
			hashCH <- m
		}
	}
}

func (s *search) searchByHash(phashCH chan<- *Media, hashCH <-chan *Media, foundCH chan<- *Media) {
	defer func() {
		close(phashCH)
		log.Println("[done] searchByHash")
	}()

	for m := range hashCH {
		if work("hash", m) {
			foundCH <- m
		} else {
			phashCH <- m
		}
	}
}

func (s *search) searchByPerceptualHash(phashCH <-chan *Media, foundCH chan<- *Media) {
	defer log.Println("[done] searchByPerceptualHash")

	for m := range phashCH {
		if work("p-hash", m) {
			foundCH <- m
		}
	}
}

func work(t string, m *Media) bool {
	fmt.Printf("> [%s] %s\n", t, m.Name)
	time.Sleep(time.Millisecond * 100)
	return true
}
