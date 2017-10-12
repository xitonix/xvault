package obfuscate

import (
	"testing"
	"time"
)

func TestPipeClosure(t *testing.T) {
	pipe := NewPipe(10)
	closed := false
	pipe.open()
	go func(closed *bool) {
		<-pipe.ch
		*closed = true
	}(&closed)

	pipe.shutdown()
	time.Sleep(2 * time.Millisecond)

	if !closed {
		t.Error("The pipe channel was supposed to be closed")
	}
}

func TestPipeClosureWithTaps(t *testing.T) {
	tap1 := newMockedTap(false, 0)
	tap2 := newMockedTap(false, 0)
	pipe := NewPipe(10)
	pipe.attachTaps(tap1, tap2)
	closed := false
	pipe.open()
	go func(closed *bool) {
		<-pipe.ch
		*closed = true
	}(&closed)

	if len(pipe.taps) != 2 {
		t.Errorf("expected number of taps was 2, actual was %d", len(pipe.taps))
	}

	for _, tap := range pipe.taps {
		if !tap.IsOpen() {
			t.Error("All the taps were supposed to be open")
		}
	}

	pipe.shutdown()

	for _, tap := range pipe.taps {
		if tap.IsOpen() {
			t.Error("All the taps were supposed to be closed")
		}
	}

	time.Sleep(2 * time.Millisecond)

	if !closed {
		t.Error("The pipe channel was supposed to be closed")
	}
}

func TestUnitOfWorkFlowToThePipe(t *testing.T) {
	testCases := []struct {
		title                  string
		numberOfWorkUnits      int
		expectedUnitsInThePipe int
		openThePipe            bool
	}{
		{
			title:                  "nothing_to_push_after_opening_the_tap_with_closed_pipe",
			numberOfWorkUnits:      0,
			expectedUnitsInThePipe: 0,
		},
		{
			title:                  "with_some_work_units_closed_pipe",
			numberOfWorkUnits:      10,
			expectedUnitsInThePipe: 0,
		},
		{
			title:                  "nothing_to_push_after_opening_the_tap_with_opened_pipe",
			numberOfWorkUnits:      0,
			openThePipe:            true,
			expectedUnitsInThePipe: 0,
		},
		{
			title:                  "with_some_work_units_opened_pipe",
			numberOfWorkUnits:      10,
			openThePipe:            true,
			expectedUnitsInThePipe: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			tap := newMockedTap(false, tc.numberOfWorkUnits)
			pipe := NewPipe(10)
			if tc.openThePipe {
				pipe.open()
			}
			var count int
			go func() {
				for range pipe.ch {
					count++
				}
			}()
			tap.Connect(pipe)
			tap.Open()
			time.Sleep(5 * time.Millisecond)
			tap.Close()
			pipe.shutdown()
			time.Sleep(5 * time.Millisecond)

			if tc.expectedUnitsInThePipe != count {
				t.Errorf("Expected to receive %d work units, but received %d", tc.expectedUnitsInThePipe, count)
			}
		})
	}
}

func TestUnitOfWorkFlowAfterOpeningThePipe(t *testing.T) {
	const numberOfWorkUnits = 10
	tap := newMockedTap(false, numberOfWorkUnits)
	pipe := NewPipe(10)

	var count int
	go func() {
		for range pipe.ch {
			count++
		}
	}()
	tap.Connect(pipe)
	tap.Open()
	time.Sleep(5 * time.Millisecond)
	if count > 0 {
		t.Errorf("No work units should have flowed into a closed pipe, but received %d units", count)
	}
	pipe.open()
	time.Sleep(5 * time.Millisecond)

	if numberOfWorkUnits != count {
		t.Errorf("Expected to receive %d work units to flow into pipe, but received %d", numberOfWorkUnits, count)
	}

	pipe.shutdown()
	tap.Close()
}
