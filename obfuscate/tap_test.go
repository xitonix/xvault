package obfuscate

import (
	"testing"

	"github.com/xitonix/xvault/assert"
)

func TestPushFromAnUnboundTap(t *testing.T) {
	tap := newMockedTap(false, 1)
	err := tap.pipe.Push(&WorkUnit{})
	if err != ErrClosedTap {
		t.Errorf("expected to get '%v' as error, but received '%v'", ErrClosedTap, err)
	}
}

func TestPushFromABoundOpenTap(t *testing.T) {
	tap := newMockedTap(false, 1)
	pipe := NewPipe(10)
	tap.Connect(pipe)
	tap.Open()
	key, err := KeyFromPassword("password")
	assert.Errors(t, false, err, nil)

	_, err = NewBucket(key, pipe, tap)
	assert.Errors(t, false, err, nil)

	err = tap.pipe.Push(&WorkUnit{})

	if err != nil {
		t.Errorf("expected to get 'nil' as error, but received '%v'", err)
	}
}

func TestPushFromABoundOpenTapWithoutBucket(t *testing.T) {
	tap := newMockedTap(false, 1)
	pipe := NewPipe(10)
	tap.Connect(pipe)
	tap.Open()

	err := tap.pipe.Push(&WorkUnit{})

	if err != nil {
		t.Errorf("expected to get 'nil' as error, but received '%v'", err)
	}
}
