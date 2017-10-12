package obfuscate

import (
	"io"
	"sync"
)

// Operation represents the operation which needs to be done by a task
type Operation int8

const (
	// Encode encryption mode
	Encode Operation = iota
	// Decode decryption mode
	Decode
)

// Task is a unit of encryption/decryption work
type Task struct {
	name  string
	mode  Operation
	input io.Reader

	// Error the error details of a failed task
	Error error

	status Status
	// MetaData key-value store of custom task data
	MetaData map[string]interface{}

	mux        sync.Mutex
	inProgress bool
	outputs    []io.Writer
}

// Name returns the name of the task
func (t *Task) Name() string {
	return t.name
}

// NewTask creates a new task object
func NewTask(name string, mode Operation, input io.Reader, output io.Writer) *Task {
	return &Task{
		name:     name,
		mode:     mode,
		input:    input,
		outputs:  []io.Writer{output},
		MetaData: make(map[string]interface{}),
		status:   Queued,
	}
}

// AddMetadata adds metadata to the
func (t *Task) AddMetadata(key string, value interface{}) {
	t.MetaData[key] = value
}

// AddOutput adds a new new output to the task
// Calling this function on an in-progress task will return ErrOperationInProgress error
// You can check the progress state of a task by calling IsRunning method
func (t *Task) AddOutput(output io.Writer) error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.inProgress {
		return ErrOperationInProgress
	}
	t.outputs = append(t.outputs, output)
	return nil
}

// CloseInput closes the input Reader.
// If the reader is not a io.Closer, calling this function will have no effect
// Calling this function on an in-progress task will return ErrOperationInProgress error
func (t *Task) CloseInput() error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.inProgress {
		return ErrOperationInProgress
	}
	input, ok := t.input.(io.Closer)
	if ok && input != nil {
		return input.Close()
	}
	return nil
}

// CloseOutputs closes all the output Writers.
// If the output is not a io.Closer, calling this function will have no effect
// Calling this function on an in-progress task will return ErrOperationInProgress error
func (t *Task) CloseOutputs() error {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.inProgress {
		return ErrOperationInProgress
	}
	for _, out := range t.outputs {
		output, ok := out.(io.Closer)
		if ok && output != nil {
			err := output.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Task) markAsInProgress() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.inProgress = true
}

func (t *Task) Status() Status {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.status
}

func (t *Task) markAsComplete(status Status) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.status = status
	t.inProgress = false
}
