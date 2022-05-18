package app

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
	filterCh chan media.Media
	filterFn func(m media.Media) bool
	mediaCh  chan media.Media
	matchCh  chan media.Media
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
		logCh:    make(chan Log, backpressure),
		errorCh:  make(chan error, backpressure),
		mediaCh:  make(chan media.Media, backpressure),
		filterCh: make(chan media.Media, backpressure),
		matchCh:  make(chan media.Media, backpressure),
	}
	fs.init()

	return fs
}

func (fs *FileSearchStream) init() {
	// handle log.
	fs.wg.Add(1)
	go func() {
		defer func() {
			//log.Println("\t[x] OnLog()")
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

	// handle filter.
	fs.wg.Add(1)
	go func() {
		defer func() {
			//log.Println("\t[x] OnFilter()")
			fs.wg.Done()
		}()

		for m := range fs.mediaCh {
			if fs.filterFn == nil {
				fs.filterCh <- m
				continue
			}

			if fs.filterFn(m) {
				fs.filterCh <- m
			}
		}

		close(fs.filterCh)
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

func (fs *FileSearchStream) OnFilter(fn func(m media.Media) bool) *FileSearchStream {
	fs.filterFn = fn
	return fs
}

func (fs *FileSearchStream) OnSearch(fn FileSearchFunc) *FileSearchStream {
	var ch chan media.Media

	if fs.pipes.len == 0 {
		ch = fs.filterCh
	} else {
		ch = make(chan media.Media, backpressure)
	}

	fs.pipes.insert(pipe{ch: ch, fn: fn})

	return fs
}

func (fs *FileSearchStream) OnMatch(fn func(m media.Media)) {
	//defer log.Println("\t[x] OnMatch()")
	fs.initSearch()
	go fs.runWalkDir()

	for m := range fs.matchCh {
		fn(m)
	}

	close(fs.logCh)
	close(fs.errorCh)
	fs.wg.Wait()
}

func (fs *FileSearchStream) Stop() {
	fs.cancelFn()
	close(fs.matchCh)
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

func (fs *FileSearchStream) runSearch(fn FileSearchFunc, outCh chan<- media.Media, inCh <-chan media.Media) {
	//defer log.Println("\t[x] runSearch()")

	for m := range inCh {
		match, err := fn(fs.ctx, m)
		if err != nil {
			fs.errorCh <- err
			continue
		}
		if match {
			fs.matchCh <- m
		} else if outCh != nil {
			outCh <- m
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
				fs.logCh <- Log{
					FileName: m.Name,
					FileType: m.ContentType,
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
