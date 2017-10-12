package obfuscate

import "sync"

type Pipe struct {
	ch   chan *WorkUnit
	taps []Tap

	mux    sync.Mutex
	isOpen bool
}

func NewPipe(bufferSize uint16) *Pipe {
	return &Pipe{
		ch: make(chan *WorkUnit, bufferSize),
	}
}

func (p *Pipe) attachTaps(tap ...Tap) {
	p.taps = append(p.taps, tap...)
}

func (p *Pipe) Push(wu *WorkUnit) {
	//even if the pipe is not open we still accept requests
	//they will get blocked until the processor starts to consume them
	p.ch <- wu
}

func (p *Pipe) close() {
	p.mux.Lock()
	defer p.mux.Unlock()
	if !p.isOpen {
		return
	}
	for _, tap := range p.taps {
		if tap.IsOpen() {
			tap.Close()
		}
	}
	close(p.ch)
	p.isOpen = false
}

func (p *Pipe) open() {
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.isOpen {
		return
	}

	for _, tap := range p.taps {
		if !tap.IsOpen() {
			tap.Open(p)
		}
	}
	p.isOpen = true
}
