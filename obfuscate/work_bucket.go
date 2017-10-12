package obfuscate

import (
	"context"
	"sync"
)

// WorkBucket is the type that implements the functionality of encrypting and decrypting
// a vault's content.
type WorkBucket struct {
	pipe   *Pipe
	master *MasterKey
	wg     *sync.WaitGroup
	cancel context.CancelFunc

	mux    sync.Mutex
	isOpen bool
}

// NewBucket creates a new instance of a vault processor object
func NewBucket(master *MasterKey, pipe *Pipe, taps ...Tap) (*WorkBucket, error) {
	pipe.attachTaps(taps...)
	tp := &WorkBucket{
		pipe:   pipe,
		master: master,
		wg:     &sync.WaitGroup{},
	}

	return tp, nil
}

// Open starts the vault processor to serve the requests coming through over
// the vault's WorkList channel. Once you are finished with the processor, you need to
// call the Close function. It's safe to call this method on a running processor
func (w *WorkBucket) Open() {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.isOpen {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	// if w.workList is not a buffered channel, cap(w.workList) will be zero
	for i := 0; i <= cap(w.pipe.ch); i++ {
		w.wg.Add(1)
		go w.readFromThePipe(ctx)
	}
	w.pipe.open()
	w.isOpen = true
}

// Close stops the processor and releases the resources.
// It's safe to call this function on a stopped processor
func (w *WorkBucket) Close() {
	w.mux.Lock()
	defer w.mux.Unlock()

	if w.cancel != nil && w.isOpen {
		w.pipe.close()
		w.cancel()
		w.wg.Wait()
		w.isOpen = false
	}
}

func (w *WorkBucket) readFromThePipe(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case wu, more := <-w.pipe.ch:
			if !more {
				return
			}
			processTask(ctx, wu.task, w.master)
			if wu.callback != nil {
				wu.callBack()
			}
		case <-ctx.Done():
			return
		}
	}
}

func processTask(ctx context.Context, task *Task, master *MasterKey) {
	task.markAsInProgress()
	var status Status
	if task.mode == Encode {
		encoder := NewEncoder(defaultBufferSize, master, task.input, task.outputs...)
		status, task.Error = encoder.EncodeContext(ctx)
	} else {
		encoder := NewDecoder(defaultBufferSize, master, task.input, task.outputs...)
		status, task.Error = encoder.DecodeContext(ctx)
	}
	task.markAsComplete(status)
}
