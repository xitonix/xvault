package obfuscate

import (
	"testing"
	"time"
)

func TestClosure(t *testing.T) {
	tap := newMockedTap()
	stream := newStream(1, tap)
	closed := false
	stream.open()
	go func(closed *bool) {
		<-stream.workList
		*closed = true
	}(&closed)

	if !tap.IsOpen() {
		t.Error("The tap was supposed to be open")
	}

	stream.shutdown()
	time.Sleep(2 * time.Millisecond)

	if !closed {
		t.Error("The stream was supposed to be closed")
	}

	if tap.IsOpen() {
		t.Error("The tap was supposed to be closed")
	}
}
