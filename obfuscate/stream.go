package obfuscate

import (
	"sync"
)

type WorkList chan *WorkUnit

type stream struct {
	workList WorkList
	inputTap Tap

	wg sync.WaitGroup
	// to stop processing the work units
	// event if the tap is still sending requests
	done chan None

	openOnce     sync.Once
	shutdownOnce sync.Once

	// to prevent multiple go routines to run
	// shutdown and open at the same time
	mux    sync.Mutex
	opened bool
}

func newStream(bufferSize uint16, tap Tap) *stream {
	return &stream{
		workList: make(WorkList, bufferSize),
		done:     make(chan None),
		inputTap: tap,
	}
}

func (s *stream) open() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.opened {
		return
	}

	s.openOnce.Do(func() {
		s.wg.Add(1)
		go s.consumeTap()
		if !s.inputTap.IsOpen() {
			s.inputTap.Open()
		}
		s.opened = true
	})
}

func (s *stream) consumeTap() {
	defer s.wg.Done()
	for {
		select {
		case <-s.done:
			return
		case w, more := <-s.inputTap.Pipe():
			if !more {
				return
			}
			s.workList <- w
		}
	}
}

func (s *stream) shutdown() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.opened {
		return
	}

	s.shutdownOnce.Do(func() {
		if s.inputTap.IsOpen() {
			s.inputTap.Close()
		}
		// stop processing the work units event if
		// the tap is still sending requests after it's closed
		close(s.done)
		s.wg.Wait()
		// Signal the engine that we are done
		close(s.workList)
		s.opened = false
	})
}
