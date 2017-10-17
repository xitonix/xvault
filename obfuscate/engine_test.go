package obfuscate

import (
	"bytes"
	"testing"
	"time"

	"github.com/mattetti/filebuffer"
)

func TestStartStop(t *testing.T) {
	tap := newMockedTap()
	engine := NewEngine(1, tap)
	engine.Start()

	if !tap.IsOpen() {
		t.Error("The tap was supposed to be open")
	}

	if !engine.IsON() {
		t.Error("The engine was supposed to be running")
	}

	engine.Stop()

	if tap.IsOpen() {
		t.Error("The tap was supposed to be closed")
	}

	if engine.IsON() {
		t.Error("The engine was supposed to be off")
	}
}

func TestInvalidMasterKey(t *testing.T) {
	var count int
	cb := func(w *WorkUnit) {
		count++
		if w.Error != errInvalidKey {
			t.Errorf("expected '%v' as error, but received '%v", errInvalidKey, w.Error)
		}
		status := w.Task.Status()
		if status != Failed {
			t.Errorf("expected status '%v', actual '%v'", Failed, status)
		}
	}

	testCases := []struct {
		title  string
		master *MasterKey
	}{
		{
			title:  "nil_master_key",
			master: nil,
		},
		{
			title:  "empty_master_key",
			master: &MasterKey{},
		},
		{
			title: "master_key_with_valid_signature",
			master: &MasterKey{
				signature: make([]byte, signatureLength),
			},
		},
		{
			title: "master_key_with_valid_signature_and_key",
			master: &MasterKey{
				signature: make([]byte, signatureLength),
				key:       make([]byte, keyLength),
			},
		},
		{
			title: "master_key_with_invalid_signature_size",
			master: &MasterKey{
				signature: make([]byte, 10),
				key:       make([]byte, keyLength),
				password:  make([]byte, 1),
			},
		},
		{
			title: "master_key_with_invalid_key_size",
			master: &MasterKey{
				signature: make([]byte, signatureLength),
				key:       make([]byte, 30),
				password:  make([]byte, 1),
			},
		},
	}

	tap := newMockedTap()
	engine := NewEngine(1, tap)
	engine.Start()
	in := filebuffer.New([]byte("input"))
	out := filebuffer.New(nil)

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			task := NewTask(Encode, in, out)
			tap.Push(NewWorkUnit(task, tc.master, cb))

			task = NewTask(Decode, out, in)
			tap.Push(NewWorkUnit(task, tc.master, cb))
		})
	}

	time.Sleep(1 * time.Millisecond)

	if count != len(testCases)*2 {
		t.Errorf("the callback function was supposed to get called %d times, but it was called %d time(s)", len(testCases)*2, count)
	}

	engine.Stop()
}

func TestEncDec(t *testing.T) {
	var count int
	cb := func(w *WorkUnit) {
		w.Task.CloseOutputs()
		w.Task.CloseInput()
		count++
		if w.Error != nil {
			t.Errorf("expected 'nil' as error, but received '%v", w.Error)
		}
		status := w.Task.Status()
		if status != Completed {
			t.Errorf("expected status '%v', actual '%v'", Completed, status)
		}
	}

	testCases := []struct {
		title string
		input []byte
	}{
		{
			title: "non_empty_input",
			input: []byte("input"),
		},
		{
			title: "empty_input",
			input: []byte(""),
		},
		{
			title: "whitespace_input",
			input: []byte("    "),
		},
	}

	tap := newMockedTap()
	engine := NewEngine(1, tap)
	engine.Start()

	master, _ := KeyFromPassword("password")

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			in := filebuffer.New(tc.input)
			out := filebuffer.New(nil)

			task := NewTask(Encode, in, out)
			tap.Push(NewWorkUnit(task, master, cb))

			//wait until the request is served
			time.Sleep(1 * time.Millisecond)

			encoded := out.Buff.Bytes()

			if len(encoded) == 0 {
				t.Errorf("encoded result is empty")
			}

			in = filebuffer.New(encoded)
			out = filebuffer.New(nil)

			task = NewTask(Decode, in, out)
			tap.Push(NewWorkUnit(task, master, cb))

			//wait until the request is served
			time.Sleep(1 * time.Millisecond)

			if !bytes.Equal(tc.input, out.Buff.Bytes()) {
				t.Errorf("decoded result does not match the input")
			}
		})
	}

	time.Sleep(1 * time.Millisecond)

	if count != len(testCases)*2 {
		t.Errorf("the callback function was supposed to get called %d times, but it was called %d time(s)", len(testCases)*2, count)
	}

	engine.Stop()
}
