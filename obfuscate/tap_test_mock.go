package obfuscate

type mockedTap struct {
	pipe              WorkList
	isOpen            bool
	numberOfWorkUnits int
	callbackCalled    bool
	withCallback      bool
	master            *MasterKey
}

func newMockedTap(withCallback bool, numberOfWorkUnits int, master *MasterKey) *mockedTap {
	return &mockedTap{
		withCallback:      withCallback,
		numberOfWorkUnits: numberOfWorkUnits,
		master:            master,
	}
}

func (m *mockedTap) IsOpen() bool {
	return m.isOpen
}

func (m *mockedTap) Channel() WorkList {
	return m.pipe
}

func (m *mockedTap) Open() {

	m.isOpen = true
	//var cb CallbackFunc
	//if m.withCallback {
	//	cb = m.callback
	//}
	//for i := 0; i < m.numberOfWorkUnits; i++ {
	//	m.pipe <- NewWorkUnit(&Task{}, m.master, cb)
	//}
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
