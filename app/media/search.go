package media

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/tsmweb/chasam/app/hash"
	"github.com/tsmweb/chasam/common/mediautil"
)

type OnError func(ctx context.Context, err error)
type OnSearch func(ctx context.Context, m *Media) (bool, error)
type OnMatch func(ctx context.Context, m *Media)

type Search struct {
	ctx       context.Context
	root      string
	hashTypes []hash.Type

	semaphoreCh chan struct{}
	errorCh     chan error
	mediaCh     chan *Media
	matchCh     chan *Media

	onError  OnError
	onSearch OnSearch
	onMatch  OnMatch
}

func NewSearch(
	ctx context.Context,
	root string,
	hashTypes []hash.Type,
	onError OnError,
	onSearch OnSearch,
	onMatch OnMatch,
	poolSize int,
) *Search {
	searchMedia := &Search{
		ctx:         ctx,
		root:        root,
		hashTypes:   hashTypes,
		semaphoreCh: make(chan struct{}, poolSize),
		errorCh:     make(chan error),
		mediaCh:     make(chan *Media),
		matchCh:     make(chan *Media),
		onError:     onError,
		onSearch:    onSearch,
		onMatch:     onMatch,
	}

	return searchMedia
}

func (s *Search) Run() {
	var wg sync.WaitGroup

	wg.Add(1)
	go s.processMedia(&wg)

	s.walkRoot(s.root)

	wg.Wait()
}

func (s *Search) processMedia(wg *sync.WaitGroup) {
	defer wg.Done()

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var wgi sync.WaitGroup

		for m := range s.mediaCh {
			wgi.Add(1)
			go s.handleSearch(m, &wgi)
		}

		wgi.Wait()
		close(s.matchCh)
	}(wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var wgi sync.WaitGroup

		for m := range s.matchCh {
			wgi.Add(1)
			go s.handleMatch(m, &wgi)
		}

		wgi.Wait()
		close(s.errorCh)
	}(wg)

	for err := range s.errorCh {
		wg.Add(1)
		go s.handleError(err, wg)
	}
}

func (s *Search) walkRoot(root string) {
	var wg sync.WaitGroup

	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		select {
		case s.semaphoreCh <- struct{}{}: // acquire token
		case <-s.ctx.Done():
			return filepath.SkipDir
		}

		wg.Add(1)
		go s.handleMedia(path, &wg)

		return nil
	})

	wg.Wait()
	close(s.mediaCh)
}

func (s *Search) handleMedia(path string, wg *sync.WaitGroup) {
	defer wg.Done()

	m, err := NewMedia(path, s.hashTypes)
	if err != nil {
		if !errors.Is(err, mediautil.ErrUnsupportedMediaType) {
			s.errorCh <- err
		}
	} else {
		s.mediaCh <- m
	}
}

func (s *Search) handleSearch(m *Media, wg *sync.WaitGroup) {
	defer wg.Done()

	ok, err := s.onSearch(s.ctx, m)
	if err != nil {
		s.errorCh <- err
	}
	if ok {
		s.matchCh <- m
	} else {
		<-s.semaphoreCh // release token
	}
}

func (s *Search) handleMatch(m *Media, wg *sync.WaitGroup) {
	defer wg.Done()

	s.onMatch(s.ctx, m)
	<-s.semaphoreCh // release token
}

func (s *Search) handleError(err error, wg *sync.WaitGroup) {
	defer wg.Done()

	s.onError(s.ctx, err)
	<-s.semaphoreCh // release token
}
