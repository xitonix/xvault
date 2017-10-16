package obfuscate

import (
	"context"
	"sync"
)

// Engine is the type that implements the functionality of encrypting and decrypting
// a vault's content.
type Engine struct {
	stream     *stream
	wg         sync.WaitGroup
	cancel     context.CancelFunc
	bufferSize uint16

	startOnce sync.Once
	stopOnce  sync.Once

	// to prevent multiple go routines to run
	// Start and Stop at the same time
	mux       sync.Mutex
	isRunning bool
}

// NewEngine creates a new instance of a vault processor object
func NewEngine(bufferSize uint16, tap Tap) *Engine {
	return &Engine{
		stream:     newStream(bufferSize, tap),
		bufferSize: bufferSize,
	}
}

// Start starts the vault processor to serve the requests coming through over
// the vault's WorkList channel. Once you are finished with the processor, you need to
// call the Stop function. It's safe to call this method on a running processor
func (e *Engine) Start() {
	e.mux.Lock()
	defer e.mux.Unlock()

	if e.isRunning {
		return
	}

	e.startOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		e.cancel = cancel

		for i := 0; uint16(i) < e.bufferSize; i++ {
			e.wg.Add(1)
			go e.monitorStream(ctx)
		}
		e.stream.open()
		e.isRunning = true
	})
}

// Stop stops the processor and releases the resources.
// It's safe to call this function on a stopped processor
func (e *Engine) Stop() {
	e.mux.Lock()
	defer e.mux.Unlock()

	if !e.isRunning {
		return
	}
	e.stopOnce.Do(func() {
		if e.cancel != nil {
			e.isRunning = false
			e.stream.shutdown()
			e.cancel()
			e.wg.Wait()
		}
	})
}

func (e *Engine) IsON() bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.isRunning
}

func (e *Engine) monitorStream(ctx context.Context) {
	defer e.wg.Done()
	for {
		select {
		case wu, more := <-e.stream.workList:
			if !more {
				return
			}

			processTask(ctx, wu)
			wu.callBack()
		case <-ctx.Done():
			return
		}
	}
}

func processTask(ctx context.Context, wu *WorkUnit) {
	wu.Task.markAsInProgress()
	var status Status
	if wu.Task.mode == Encode {
		encoder := NewEncoder(defaultBufferSize, wu.master, wu.Task.input, wu.Task.outputs...)
		status, wu.Error = encoder.EncodeContext(ctx)
	} else {
		encoder := NewDecoder(defaultBufferSize, wu.master, wu.Task.input, wu.Task.outputs...)
		status, wu.Error = encoder.DecodeContext(ctx)
	}
	wu.Task.markAsComplete(status)
}
