package fstream

import (
	"context"
	"errors"
	"fmt"
	"github.com/tsmweb/chasam/app/media"
	"github.com/tsmweb/chasam/common/mediautil"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

//var (
//	semaphoreMax = runtime.NumCPU() * 2
//)

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

type mediaSearchStream struct {
	ctx      context.Context
	dir      string
	pipes    *pipeList
	errorCh  chan error
	errorFn  func(err error)
	mediaCh  chan *media.Media
	matchCh  chan *media.Media
	globalWG sync.WaitGroup
}

func (s *mediaSearchStream) init() {
	// handle error.
	s.globalWG.Add(1)
	go func() {
		defer s.globalWG.Done()

		for err := range s.errorCh {
			if s.errorFn != nil {
				s.errorFn(err)
			} else {
				fmt.Fprint(os.Stderr, err)
			}
		}
	}()
}

func (s *mediaSearchStream) onError(fn func(err error)) *mediaSearchStream {
	s.errorFn = fn
	return s
}

func (s *mediaSearchStream) onEach(fn FileSearchFunc) *mediaSearchStream {
	var ch chan *media.Media

	if s.pipes.len == 0 {
		ch = s.mediaCh
	} else {
		ch = make(chan *media.Media)
	}

	s.pipes.insert(pipe{ch: ch, fn: fn})

	return s
}

func (s *mediaSearchStream) onMatch(fn func(m *media.Media)) {
	s.initSearch()
	go s.walkDir()

	for m := range s.matchCh {
		fn(m)
	}

	close(s.errorCh)
	s.globalWG.Wait()
}

func (s *mediaSearchStream) initSearch() {
	n := s.pipes.head

	for n != nil {
		p := n.value
		n = n.next

		if n == nil {
			go s.runSearch(p.fn, nil, p.ch)
		} else {
			go s.runSearch(p.fn, n.value.ch, p.ch)
		}
	}
}

func (s *mediaSearchStream) runSearch(fn FileSearchFunc, outCh chan<- *media.Media, inCh <-chan *media.Media) {
loop:
	for m := range inCh {
		select {
		case <-s.ctx.Done():
			break loop
		default:
			res, err := fn(s.ctx, m)
			if err != nil {
				s.errorCh <- err
				return
			}

			switch res {
			case Skip:
				return
			case Match:
				s.matchCh <- m
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
		close(s.matchCh)
	}
}

func (s *mediaSearchStream) walkDir() {
	f, err := os.Open(s.dir)
	if err != nil {
		s.errorCh <- fmt.Errorf("mediaSearchStream::walkDir(%s) | Error: %v", s.dir, err)
		return
	}
	defer f.Close()

	entries, err := f.Readdir(-1)
	if err != nil {
		s.errorCh <- fmt.Errorf("mediaSearchStream::walkDir(%s) | Error: %v", s.dir, err)
	}

	for _, entry := range entries {
		if s.cancelled() {
			break
		}

		if entry.IsDir() {
			continue
		}

		m, err := media.New(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			if !errors.Is(err, mediautil.ErrUnsupportedMediaType) {
				s.errorCh <- err
			}
		} else {
			s.mediaCh <- m
		}
	}

	close(s.mediaCh)
}

func (s *mediaSearchStream) cancelled() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

type MediaSearchStream interface {
	OnError(fn func(err error)) MediaSearchStream
	OnEach(fn FileSearchFunc) MediaSearchStream
	OnMatch(fn func(m *media.Media)) MediaSearchStream
	Run()
}

type mediaSearchStreamBuilder struct {
	ctx     context.Context
	semaCh  chan struct{}
	roots   []string
	errorFn func(err error)
	eachFn  []FileSearchFunc
	matchFn func(m *media.Media)
}

func NewMediaSearchStream(ctx context.Context, roots []string, cpu int) MediaSearchStream {
	return &mediaSearchStreamBuilder{
		ctx:    ctx,
		semaCh: make(chan struct{}, cpu),
		roots:  roots,
	}
}

func (b *mediaSearchStreamBuilder) OnError(fn func(err error)) MediaSearchStream {
	b.errorFn = fn
	return b
}

func (b *mediaSearchStreamBuilder) OnEach(fn FileSearchFunc) MediaSearchStream {
	b.eachFn = append(b.eachFn, fn)
	return b
}

func (b *mediaSearchStreamBuilder) OnMatch(fn func(m *media.Media)) MediaSearchStream {
	b.matchFn = fn
	return b
}

func (b *mediaSearchStreamBuilder) Run() {
	var wg sync.WaitGroup

	for _, root := range b.roots {
		wg.Add(1)
		root := root
		go b.walkRoot(root, &wg)
	}

	wg.Wait()
}

func (b *mediaSearchStreamBuilder) walkRoot(root string, wg *sync.WaitGroup) {
	defer wg.Done()

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			wg.Add(1)
			go b.searchMedia(path, wg)
		}
		return nil
	})
}

func (b *mediaSearchStreamBuilder) searchMedia(dir string, wg *sync.WaitGroup) {
	select {
	case b.semaCh <- struct{}{}: // acquire token
	case <-b.ctx.Done():
		return
	}

	defer func() {
		wg.Done()
		<-b.semaCh // release token
	}()

	ms := &mediaSearchStream{
		ctx:     b.ctx,
		dir:     dir,
		pipes:   new(pipeList),
		errorCh: make(chan error),
		mediaCh: make(chan *media.Media),
		matchCh: make(chan *media.Media),
	}
	ms.init()
	ms.onError(b.errorFn)
	for _, fn := range b.eachFn {
		ms.onEach(fn)
	}
	ms.onMatch(b.matchFn)
}
