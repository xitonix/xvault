package obfuscate

import (
	"context"
	"sync"
)

// Engine is the type that processes an stream of encrypt/decrypt work units.
//
// Once the processing of a work unit has been finished, the engine will call
// the CallbackFunc of the unit (if specified) and sends the processing result back
// to the function.
//
// In order to feed the engine with work units, you need to connect
// your implementation of the Tap interface to it
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

// NewEngine creates a new instance of the Engine type.
func NewEngine(bufferSize uint16, tap Tap) *Engine {
	return &Engine{
		stream:     newStream(bufferSize, tap),
		bufferSize: bufferSize,
	}
}

// Start starts processing the work unit stream provided by the input Tap.
// Once you are finished with the Engine, you need to call the Stop function.
//
// Starting the engine automatically opens the input tap. You SHOULD NOT
// call the tap's Open function explicitly.
//
// NOTE: You can only call the Start method once.
// Restarting an engine object has no effect.
func (e *Engine) Start() {
	e.mux.Lock()
	defer e.mux.Unlock()

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

// Stop stops the engine and releases the resources.
// Stopping the engine will automatically close the input tap, so you don't need to
// explicitly call the tap's Close function.
//
// NOTE: Once the engine has been stopped, starting it will have no effect.
func (e *Engine) Stop() {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.stopOnce.Do(func() {
		if e.cancel != nil {
			e.isRunning = false
			e.stream.shutdown()
			e.cancel()
			e.wg.Wait()
		}
	})
}

// IsON returns true if the engine has been started, otherwise returns false.
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
