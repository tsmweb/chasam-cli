package fstream

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/common/mediautil"
	"os"
	"path/filepath"
	"sync"
)

const (
	backpressure = 128 // 128
	semaphoreMax = 20  // 20
)

type ResultType int

const (
	Match ResultType = iota
	Next
	Skip
)

type FileSearchFunc func(ctx context.Context, m *media.Media) (ResultType, error)

type pipe struct {
	ch chan *media.Media
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
	ctx     context.Context
	roots   []string
	pipes   *pipeList
	semaCh  chan struct{}
	errorCh chan error
	errorFn func(err error)
	mediaCh chan *media.Media
	matchCh chan *media.Media
	wg      sync.WaitGroup
}

func NewFileSearchStream(ctx context.Context, roots []string) *FileSearchStream {
	fs := &FileSearchStream{
		ctx:     ctx,
		roots:   roots,
		pipes:   new(pipeList),
		semaCh:  make(chan struct{}, semaphoreMax),
		errorCh: make(chan error, backpressure),
		mediaCh: make(chan *media.Media, backpressure),
		matchCh: make(chan *media.Media, backpressure),
	}
	fs.init()

	return fs
}

func (fs *FileSearchStream) init() {
	// handle error.
	fs.wg.Add(1)
	go func() {
		defer func() {
			//log.Println("\t[x] OnError()")
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

func (fs *FileSearchStream) OnError(fn func(err error)) *FileSearchStream {
	fs.errorFn = fn
	return fs
}

func (fs *FileSearchStream) OnEach(fn FileSearchFunc) *FileSearchStream {
	var ch chan *media.Media

	if fs.pipes.len == 0 {
		ch = fs.mediaCh
	} else {
		ch = make(chan *media.Media, backpressure)
	}

	fs.pipes.insert(pipe{ch: ch, fn: fn})

	return fs
}

func (fs *FileSearchStream) OnMatch(fn func(m *media.Media)) {
	fs.initSearch()
	go fs.runWalkDir()

	for m := range fs.matchCh {
		fn(m)
	}

	close(fs.errorCh)
	fs.wg.Wait()
}

func (fs *FileSearchStream) initSearch() {
	n := fs.pipes.head

	for n != nil {
		p := n.value
		n = n.next

		if n == nil {
			go fs.runSearch(p.fn, nil, p.ch)
		} else {
			go fs.runSearch(p.fn, n.value.ch, p.ch)
		}
	}
}

func (fs *FileSearchStream) runSearch(fn FileSearchFunc, outCh chan<- *media.Media, inCh <-chan *media.Media) {
loop:
	for m := range inCh {
		select {
		case <-fs.ctx.Done():
			break loop
		default:
			res, err := fn(fs.ctx, m)
			if err != nil {
				fs.errorCh <- err
				continue
			}

			switch res {
			case Skip:
				continue
			case Match:
				fs.matchCh <- m
			case Next:
				if outCh != nil {
					outCh <- m
				}
			}
		}
	}

	if outCh != nil {
		close(outCh)
	} else {
		close(fs.matchCh)
	}
}

func (fs *FileSearchStream) runWalkDir() {
	//defer log.Println("\t[x] runWalkDir()")
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
				fs.mediaCh <- m
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

	dirs, err := f.Readdir(-1)
	if err != nil {
		fs.errorCh <- fmt.Errorf("FileSearchStream::dirents(%s) | Error: %v", dir, err)
	}

	return dirs
}

func (fs *FileSearchStream) cancelled() bool {
	select {
	case <-fs.ctx.Done():
		return true
	default:
		return false
	}
}
