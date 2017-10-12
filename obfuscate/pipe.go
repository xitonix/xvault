package obfuscate

import "sync"

type Pipe struct {
	ch    chan *WorkUnit
	queue []*WorkUnit
	taps  []Tap

	mux    sync.Mutex
	isOpen bool
}

func NewPipe(bufferSize uint16) *Pipe {
	return &Pipe{
		ch: make(chan *WorkUnit, bufferSize),
	}
}

func (p *Pipe) attachTaps(taps ...Tap) {
	for _, tap := range taps {
		p.taps = append(p.taps, tap)
		tap.Connect(p)
	}
}

func (p *Pipe) Push(wu *WorkUnit) error {
	if p == nil {
		return ErrClosedTap
	}

	p.mux.Lock()
	defer p.mux.Unlock()
	if !p.isOpen {
		//buffer the work units if the pipe is not open
		p.queue = append(p.queue, wu)
		return nil
	}

	p.ch <- wu
	return nil
}

func (p *Pipe) shutdown() {
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
			tap.Open()
		}
	}

	p.isOpen = true
	p.flushTheQueue()
}

func (p *Pipe) flushTheQueue() {
	//In case there is no consumer reading off the channel,
	//we don't want flushing the queue to block the call to open()
	go func() {
		for _, wu := range p.queue {
			if !p.isOpen {
				break
			}
			p.ch <- wu
		}
		p.queue = nil
	}()
}
