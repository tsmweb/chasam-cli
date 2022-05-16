package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/common/mediautil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	backpressure = 128 // 128
	semaphoreMax = 20  // 20
)

type Log struct {
	FileName string
	FileType string
	FilePath string
}

type FileSearchFunc func(ctx context.Context, m media.Media) (bool, error)

type pipe struct {
	ch chan media.Media
	fn FileSearchFunc
}

type node struct {
	value pipe
	next  *node
}

type pipeList struct {
	head *node
	tail *node
	len  int
}

func (l *pipeList) insert(p pipe) {
	n := &node{value: p}

	if l.head == nil {
		l.head = n
	}
	if l.tail != nil {
		l.tail.next = n
	}

	l.tail = n
	l.len++
}

type FileSearchStream struct {
	ctx      context.Context
	cancelFn context.CancelFunc
	roots    []string
	pipes    *pipeList
	semaCh   chan struct{}
	logCh    chan Log
	logFn    func(l Log)
	errorCh  chan error
	errorFn  func(err error)
	mediaCh  chan media.Media
	foundCh  chan media.Media
	wg       sync.WaitGroup
}

func NewFileSearchStream(roots []string) *FileSearchStream {
	ctx, cancelFunc := context.WithCancel(context.Background())

	fs := &FileSearchStream{
		ctx:      ctx,
		cancelFn: cancelFunc,
		roots:    roots,
		pipes:    new(pipeList),
		semaCh:   make(chan struct{}, semaphoreMax),
		mediaCh:  make(chan media.Media, backpressure),
		logCh:    make(chan Log, backpressure),
		errorCh:  make(chan error, backpressure),
		foundCh:  make(chan media.Media, backpressure),
	}
	fs.init()

	return fs
}

func (fs *FileSearchStream) init() {
	// handle log.
	fs.wg.Add(1)
	go func() {
		defer func() {
			log.Println("\t[x] OnLog()")
			fs.wg.Done()
		}()

		for l := range fs.logCh {
			if fs.logFn != nil {
				fs.logFn(l)
			}
		}
	}()

	// handle error.
	fs.wg.Add(1)
	go func() {
		defer func() {
			log.Println("\t[x] OnError()")
			fs.wg.Done()
		}()

		for err := range fs.errorCh {
			if fs.errorFn != nil {
				fs.errorFn(err)
			} else {
				fmt.Fprint(os.Stderr, err)
			}
		}
	}()
}

func (fs *FileSearchStream) OnLog(fn func(l Log)) *FileSearchStream {
	fs.logFn = fn
	return fs
}

func (fs *FileSearchStream) OnError(fn func(err error)) *FileSearchStream {
	fs.errorFn = fn
	return fs
}

func (fs *FileSearchStream) OnPipe(fn FileSearchFunc) *FileSearchStream {
	var ch chan media.Media

	if fs.pipes.len == 0 {
		ch = fs.mediaCh
	} else {
		ch = make(chan media.Media, backpressure)
	}

	fs.pipes.insert(pipe{ch: ch, fn: fn})

	return fs
}

func (fs *FileSearchStream) OnFound(fn func(ctx context.Context, m media.Media)) {
	defer log.Println("\t[x] OnFound()")
	fs.initPipe()
	go fs.runSearch()

	for m := range fs.foundCh {
		fn(fs.ctx, m)
	}

	close(fs.logCh)
	close(fs.errorCh)
	fs.wg.Wait()
}

func (fs *FileSearchStream) Stop() {
	fs.cancelFn()
	close(fs.foundCh)
}

func (fs *FileSearchStream) initPipe() {
	n := fs.pipes.head

	for n != nil {
		p := n.value
		n = n.next

		if n == nil {
			go fs.runPipe(p.fn, nil, p.ch)
		} else {
			go fs.runPipe(p.fn, n.value.ch, p.ch)
		}
	}
}

func (fs *FileSearchStream) runPipe(fn FileSearchFunc, outCh chan<- media.Media, inCh <-chan media.Media) {
	defer func() {
		if outCh != nil {
			close(outCh)
		} else {
			close(fs.foundCh)
		}
		log.Println("\t[x] runPipe()")
	}()

	for m := range inCh {
		found, err := fn(fs.ctx, m)
		if err != nil {
			fs.errorCh <- err
			continue
		}
		if found {
			fs.foundCh <- m
		} else if outCh != nil {
			outCh <- m
		}
	}
}

func (fs *FileSearchStream) runSearch() {
	defer log.Println("\t[x] runSearch()")
	var wg sync.WaitGroup

	for _, dir := range fs.roots {
		wg.Add(1)
		go fs.walkDir(dir, &wg)
	}

	wg.Wait()
	close(fs.mediaCh)
}

func (fs *FileSearchStream) walkDir(dir string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, entry := range fs.dirents(dir) {
		if fs.cancelled() {
			break
		}

		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go fs.walkDir(subdir, wg)
		} else {
			m, err := media.New(filepath.Join(dir, entry.Name()))
			if err != nil {
				if !errors.Is(err, mediautil.ErrUnsupportedMediaType) {
					fs.errorCh <- err
				}
			} else {
				fs.logCh <- Log{
					FileName: m.Name,
					FileType: m.Type,
					FilePath: m.Path,
				}
				fs.mediaCh <- *m
			}
		}
	}
}

func (fs *FileSearchStream) dirents(dir string) []os.FileInfo {
	select {
	case fs.semaCh <- struct{}{}: // acquire token
	case <-fs.ctx.Done():
		return nil // cancelled
	}

	defer func() { <-fs.semaCh }() // release token

	f, err := os.Open(dir)
	if err != nil {
		fs.errorCh <- fmt.Errorf("FileSearchStream::dirents(%s) | Error: %v", dir, err)
		return nil
	}
	defer f.Close()

	entries, err := f.Readdir(0)
	if err != nil {
		fs.errorCh <- fmt.Errorf("FileSearchStream::dirents(%s) | Error: %v", dir, err)
	}

	return entries
}

func (fs *FileSearchStream) cancelled() bool {
	select {
	case <-fs.ctx.Done():
		return true
	default:
		return false
	}
}
