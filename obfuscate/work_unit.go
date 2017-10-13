package obfuscate

// CallbackFunc is a callback function which will get called by the bucket once
// the processing of a work unit has been finished
type CallbackFunc func(*WorkUnit)
type MetadataMap map[string]interface{}

// WorkUnit is a unit of encryption/decryption work
type WorkUnit struct {
	Task     *Task
	master   *MasterKey
	callback CallbackFunc
	Metadata MetadataMap
}

// NewWorkUnit creates a new work unit
func NewWorkUnit(t *Task, master *MasterKey, callback CallbackFunc) *WorkUnit {
	return &WorkUnit{
		Task:     t,
		master:   master,
		callback: callback,
		Metadata: make(MetadataMap),
	}
}

func (w *WorkUnit) callBack() {
	if w.callback != nil {
		w.callback(w)
	}
}
