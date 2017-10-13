package obfuscate

import (
	"sync"
)

type RequestChannel chan *WorkUnit

type stream struct {
	tube RequestChannel
	tap  Tap

	wg   *sync.WaitGroup
	done chan None

	openOnce     sync.Once
	shutdownOnce sync.Once

	//to prevent multiple go routines to run shutdown and open at the same time
	mux    sync.Mutex
	isOpen bool
}

func newStream(bufferSize uint16, tap Tap) *stream {
	return &stream{
		tube: make(RequestChannel, bufferSize),
		done: make(chan None),
		wg:   &sync.WaitGroup{},
		tap:  tap,
	}
}

func (s *stream) consumeTap() {
	defer s.wg.Done()
	for {
		select {
		case <-s.done:
			return
		case w, more := <-s.tap.Requests():
			if !more {
				return
			}
			s.tube <- w
		}
	}
}

func (s *stream) shutdown() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.isOpen {
		return
	}

	s.shutdownOnce.Do(func() {
		if s.tap.IsOpen() {
			s.tap.Close()
		}
		close(s.done)
		s.wg.Wait()
		close(s.tube)
		s.isOpen = false
	})
}

func (s *stream) open() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.isOpen {
		return
	}

	s.openOnce.Do(func() {
		s.wg.Add(1)
		go s.consumeTap()
		if !s.tap.IsOpen() {
			s.tap.Open()
		}
		s.isOpen = true
	})
}
