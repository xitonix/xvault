package obfuscate

type mockedTap struct {
	pipe              *Pipe
	isOpen            bool
	numberOfWorkUnits int
	callbackCalled    bool
	withCallback      bool
}

func newMockedTap(withCallback bool, numberOfWorkUnits int) *mockedTap {
	return &mockedTap{withCallback: withCallback, numberOfWorkUnits: numberOfWorkUnits}
}

func (m *mockedTap) IsOpen() bool {
	return m.isOpen
}

func (m *mockedTap) Connect(pipe *Pipe) {
	m.pipe = pipe
}

func (m *mockedTap) Open() {

	m.isOpen = true
	var cb CallbackFunc
	if m.withCallback {
		cb = m.callback
	}
	for i := 0; i < m.numberOfWorkUnits; i++ {
		m.pipe.Push(NewWorkUnit(&Task{}, cb))
	}
}

func (m *mockedTap) Close() {
	m.isOpen = false
}

func (m *mockedTap) callback(t *Task) {
	m.callbackCalled = true
}

func (m *mockedTap) IsCallbackCalled() bool {
	return m.callbackCalled
}
