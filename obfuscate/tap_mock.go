package obfuscate

type mockedTap struct {
	pipe              WorkList
	isOpen            bool
	callbackCalled    bool
	withCallback      bool
}

func newMockedTap() *mockedTap {
	return &mockedTap{
		pipe: make(WorkList),
	}
}

func (m *mockedTap) IsOpen() bool {
	return m.isOpen
}

func (m *mockedTap) Pipe() WorkList {
	return m.pipe
}

func (m *mockedTap) Open() {
	m.isOpen = true
}

func (m *mockedTap) Push(wUnits ...*WorkUnit) {
	for _, wu := range wUnits{
		m.pipe <- wu
	}
}

func (m *mockedTap) Close() {
	m.isOpen = false
}
