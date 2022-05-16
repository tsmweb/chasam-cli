package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/common/imageutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	backpressure = 128 // 128
	semaphoreMax = 20  // 20
)

type FileStreamFunc func(ctx context.Context, m media.Media) bool

func (f FileStreamFunc) Execute(ctx context.Context, m media.Media) bool {
	return f(ctx, m)
}

type pipe struct {
	ch chan media.Media
	fn FileStreamFunc
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

type FileStream struct {
	ctx      context.Context
	cancelFn context.CancelFunc
	roots    []string
	pipes    *pipeList
	semaCh   chan struct{}
	mediaCh  chan media.Media
	errorCh  chan error
	matchCh  chan media.Media
}

func NewFileStream(roots []string) *FileStream {
	ctx, cancelFunc := context.WithCancel(context.Background())

	fs := &FileStream{
		ctx:      ctx,
		cancelFn: cancelFunc,
		roots:    roots,
		pipes:    new(pipeList),
		semaCh:   make(chan struct{}, semaphoreMax),
		mediaCh:  make(chan media.Media, backpressure),
		errorCh:  make(chan error, backpressure),
		matchCh:  make(chan media.Media, backpressure),
	}

	return fs
}

func (fs *FileStream) OnError(fn func(err error)) *FileStream {
	done := make(chan struct{})

	go func(fn func(err error)) {
		defer log.Println("\t[done] OnError()")
		close(done)
		for err := range fs.errorCh {
			fn(err)
		}
	}(fn)

	<-done

	return fs
}

func (fs *FileStream) OnPipe(fn FileStreamFunc) *FileStream {
	var ch chan media.Media

	if fs.pipes.len == 0 {
		ch = fs.mediaCh
	} else {
		ch = make(chan media.Media, backpressure)
	}

	fs.pipes.insert(pipe{ch: ch, fn: fn})

	return fs
}

// OnMatch inicia o fluxo.
func (fs *FileStream) OnMatch(fn func(ctx context.Context, m media.Media)) {
	defer log.Println("\t[done] OnMatch()")
	fs.makePipe()
	go fs.run()

	for m := range fs.matchCh {
		fn(fs.ctx, m)
	}
}

func (fs *FileStream) Stop() {
	fs.cancelFn()
	close(fs.errorCh)
	close(fs.matchCh)
}

func (fs *FileStream) makePipe() {
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

func (fs *FileStream) runPipe(fn FileStreamFunc, outCh chan<- media.Media, inCh <-chan media.Media) {
	defer func() {
		if outCh != nil {
			close(outCh)
		} else {
			close(fs.matchCh)
		}
		log.Println("\t[done] runPipe()")
	}()

	for m := range inCh {
		if fn(fs.ctx, m) {
			fs.matchCh <- m
		} else if outCh != nil {
			outCh <- m
		}
	}
}

func (fs *FileStream) run() {
	defer log.Println("\t[done] run()")
	var wg sync.WaitGroup

	for _, dir := range fs.roots {
		wg.Add(1)
		go fs.walkDir(dir, &wg)
	}

	wg.Wait()
	close(fs.mediaCh)
	close(fs.errorCh)
}

func (fs *FileStream) walkDir(dir string, wg *sync.WaitGroup) {
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
				if !errors.Is(err, imageutil.ErrUnsupportedMediaType) {
					fs.errorCh <- err
				}
			} else {
				fs.mediaCh <- *m
			}
		}
	}
}

func (fs *FileStream) dirents(dir string) []os.FileInfo {
	select {
	case fs.semaCh <- struct{}{}: // acquire token
	case <-fs.ctx.Done():
		return nil // cancelled
	}

	defer func() { <-fs.semaCh }() // release token

	f, err := os.Open(dir)
	if err != nil {
		fs.errorCh <- fmt.Errorf("FileStream::dirents(%s) | Error: %v", dir, err)
		return nil
	}
	defer f.Close()

	entries, err := f.Readdir(0)
	if err != nil {
		fs.errorCh <- fmt.Errorf("FileStream::dirents(%s) | Error: %v", dir, err)
	}

	return entries
}

func (fs *FileStream) cancelled() bool {
	select {
	case <-fs.ctx.Done():
		return true
	default:
		return false
	}
}
