package obfuscate

import (
	"sync"
)

// WorkList is the pipe to flow the work units from the tap to the engine
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
	mux sync.Mutex
}

func newStream(bufferSize uint16, tap Tap) *stream {
	s := &stream{
		workList: make(WorkList, bufferSize),
		done:     make(chan None),
		inputTap: tap,
	}

	s.wg.Add(1)
	go s.consumeTap()
	return s
}

func (s *stream) open() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.openOnce.Do(func() {
		if s.inputTap == nil {
			return
		}
		if !s.inputTap.IsOpen() {
			s.inputTap.Open()
		}
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

	s.shutdownOnce.Do(func() {
		if s.inputTap == nil {
			return
		}
		if s.inputTap.IsOpen() {
			s.inputTap.Close()
		}
		// stop processing the work units event if
		// the tap is still sending requests after it's closed
		close(s.done)
		s.wg.Wait()
		// Signal the engine that we are done
		close(s.workList)
	})
}
