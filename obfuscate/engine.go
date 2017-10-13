package obfuscate

import (
	"context"
	"sync"
)

// Engine is the type that implements the functionality of encrypting and decrypting
// a vault's content.
type Engine struct {
	stream     *stream
	notify     bool
	progress   chan *Result
	wg         *sync.WaitGroup
	cancel     context.CancelFunc
	bufferSize uint16

	startOnce sync.Once
	stopOnce  sync.Once

	//to prevent multiple go routines to run Start and Stop at the same time
	mux       sync.Mutex
	isRunning bool
}

// NewEngine creates a new instance of a vault processor object
func NewEngine(bufferSize uint16, enableProgress bool, tap Tap) (*Engine, error) {
	return &Engine{
		stream:     newStream(bufferSize, tap),
		progress:   make(chan *Result),
		wg:         &sync.WaitGroup{},
		notify:     enableProgress,
		bufferSize: bufferSize,
	}, nil
}

func (e *Engine) Progress() <-chan *Result {
	return e.progress
}

func (e *Engine) reportProgress(r *Result) {
	if e.notify {
		e.progress <- r
	}
}

// Start starts the vault processor to serve the requests coming through over
// the vault's WorkList channel. Once you are finished with the processor, you need to
// call the Stop function. It's safe to call this method on a running processor
func (e *Engine) Start() (err error) {
	e.mux.Lock()
	defer e.mux.Unlock()

	if e.isRunning {
		return nil
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
	return
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
			close(e.progress)
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
		case wu, more := <-e.stream.tube:
			if !more {
				return
			}

			e.reportProgress(&Result{
				Status:   Queued,
				Metadata: wu.Metadata,
			})
			err := processTask(ctx, wu.Task, wu.master)
			wu.callBack()
			e.reportProgress(&Result{
				Error:    err,
				Status:   wu.Task.status,
				Metadata: wu.Metadata,
			})
		case <-ctx.Done():
			return
		}
	}
}

func processTask(ctx context.Context, task *Task, master *MasterKey) error {
	task.markAsInProgress()
	var status Status
	var err error
	if task.mode == Encode {
		encoder := NewEncoder(defaultBufferSize, master, task.input, task.outputs...)
		status, err = encoder.EncodeContext(ctx)
	} else {
		encoder := NewDecoder(defaultBufferSize, master, task.input, task.outputs...)
		status, err = encoder.DecodeContext(ctx)
	}
	task.markAsComplete(status)
	return err
}
