package obfuscate

import (
	"testing"

	"github.com/mattetti/filebuffer"
	"github.com/xitonix/xvault/obfuscate/mocks"
)

func TestTaskAddOutput(t *testing.T) {
	testCases := []struct {
		title         string
		expectedError error
		markAsRunning bool
	}{
		{
			title: "adding_output_must_append_to_the_output_slice",
		},
		{
			title:         "adding_output_to_an_in_progress_task_must_fail",
			expectedError: ErrOperationInProgress,
			markAsRunning: true,
		},
	}

	in := filebuffer.New(nil)
	out := filebuffer.New(nil)

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			task := NewTask(Encode, in, out)
			if tc.markAsRunning {
				task.markAsInProgress()
			}

			anotherOut := filebuffer.New(nil)
			err := task.AddOutput(anotherOut)
			if tc.expectedError != err {
				t.Errorf("Expected '%v' as error, but received '%v'", tc.expectedError, err)
			}

			if tc.markAsRunning {
				if len(task.outputs) != 1 {
					t.Error("There should only be one io.Writer in the output list")
				}
			} else {
				if len(task.outputs) != 2 {
					t.Error("The second io.Writer did not get added to the output list")
				}
			}
		})
	}
}

func TestTaskCloseInputOutput(t *testing.T) {
	in := &mocks.ReadCloser{}
	out := &mocks.WriteCloser{}

	task := NewTask(Encode, in, out)
	task.CloseInput()
	task.CloseOutputs()

	if !in.IsClosed {
		t.Error("Input was supposed to get closed")
	}

	if !out.IsClosed {
		t.Error("Output was supposed to get closed")
	}
}

func TestTaskCloseInputInProgress(t *testing.T) {
	in := &mocks.ReadCloser{}
	out := &mocks.WriteCloser{}

	task := NewTask(Encode, in, out)
	task.markAsInProgress()

	err := task.CloseInput()
	if ErrOperationInProgress != err {
		t.Errorf("Expected 'ErrOperationInProgress' as error, but received '%v'", err)
	}

	err = task.CloseOutputs()
	if ErrOperationInProgress != err {
		t.Errorf("Expected 'ErrOperationInProgress' as error, but received '%v'", err)
	}
}
