package obfuscate

// CallbackFunc is a callback function which will get called by the bucket once
// the processing of a work unit has been finished
type CallbackFunc func(*Task)

// WorkUnit is a unit of encryption/decryption work
type WorkUnit struct {
	task     *Task
	callback CallbackFunc
}

// NewWorkUnit creates a new work unit
func NewWorkUnit(t *Task, c CallbackFunc) *WorkUnit {
	return &WorkUnit{task: t, callback: c}
}

func (w *WorkUnit) callBack() {
	if w.callback != nil {
		w.callback(w.task)
	}
}
