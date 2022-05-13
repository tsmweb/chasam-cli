package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/common"
	"github.com/tsmweb/chasam/common/imageutil"
	"github.com/tsmweb/chasam/pkg/ebus"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ChaSAM struct {
	ctx context.Context

	chSema chan struct{} // concurrency-limiting counting semaphore
	//chMedia chan *media.Media
}

func NewChaSAM(ctx context.Context) *ChaSAM {
	app := &ChaSAM{
		ctx:    ctx,
		chSema: make(chan struct{}, 20),
	}

	return app
}

func (app *ChaSAM) SearchMedia(dirs []string) {
	var wg sync.WaitGroup

	for _, dir := range dirs {
		wg.Add(1)
		go app.walkDir(dir, &wg)
	}

	wg.Wait()
	//close(app.chMedia)
}

func (app *ChaSAM) walkDir(dir string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, entry := range app.dirents(dir) {
		if app.cancelled() {
			break
		}

		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go app.walkDir(subdir, wg)
		} else {
			m, err := media.New(filepath.Join(dir, entry.Name()))
			if err != nil {
				if !errors.Is(err, imageutil.ErrUnsupportedMediaType) {
					ebus.Instance().Publish(common.ErrorTopic, err)
				}
			} else {
				//app.chMedia <- m
				time.Sleep(time.Millisecond * 500)
				log.Printf("[>] Processando: %s\n", m.Name)
				ebus.Instance().Publish(common.MediaTopic, m)
			}
		}
	}
}

func (app *ChaSAM) dirents(dir string) []os.FileInfo {
	select {
	case app.chSema <- struct{}{}: // acquire token
	case <-app.ctx.Done():
		return nil // cancelled
	}

	defer func() { <-app.chSema }() // release token

	f, err := os.Open(dir)
	if err != nil {
		ebus.Instance().Publish(common.ErrorTopic,
			fmt.Errorf("ChaSAM::dirents(%s) | Error: %v", dir, err))
		return nil
	}
	defer f.Close()

	entries, err := f.Readdir(0)
	if err != nil {
		ebus.Instance().Publish(common.ErrorTopic, err)
	}

	return entries
}

func (app *ChaSAM) cancelled() bool {
	select {
	case <-app.ctx.Done():
		return true
	default:
		return false
	}
}
