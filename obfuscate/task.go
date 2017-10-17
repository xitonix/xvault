package obfuscate

import (
	"io"
	"sync"
)

// Operation represents the operation which needs to be done by a Task
type Operation int8

const (
	// Encode encryption mode
	Encode Operation = iota
	// Decode decryption mode
	Decode
)

// Task represents an encryption/decryption request
type Task struct {
	mode    Operation
	input   io.Reader
	status  Status
	outputs []io.Writer

	mux        sync.Mutex
	inProgress bool
}

// NewTask creates a new Task object
func NewTask(mode Operation, input io.Reader, output io.Writer) *Task {
	return &Task{
		mode:    mode,
		input:   input,
		outputs: []io.Writer{output},

		status: Queued,
	}
}

// AddOutput adds a new output to the Task
// Calling this function on an in-progress Task will return return an error of type obfuscate.ErrOperationInProgress
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
// Calling CloseInput() on an in-progress Task will return an error of type obfuscate.ErrOperationInProgress
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
// If the output is not a io.Closer, calling this function will have no effect.
//
// Calling CloseOutputs() on an in-progress Task will return an error of type obfuscate.ErrOperationInProgress
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

// Status returns the task's progress status
func (t *Task) Status() Status {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.status
}

func (t *Task) markAsInProgress() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.inProgress = true
}

func (t *Task) markAsComplete(status Status) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.status = status
	t.inProgress = false
}
