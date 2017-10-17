package obfuscate

// CallbackFunc is a callback function which will get called by the engine once
// the processing of a work unit has been finished.
type CallbackFunc func(*WorkUnit)

// MetadataMap the map of custom data
type MetadataMap map[string]interface{}

// WorkUnit is a unit of encryption/decryption work
type WorkUnit struct {
	master   *MasterKey
	callback CallbackFunc
	// Task the task which needs to be processed
	Task *Task
	// Metadata custom data
	Metadata MetadataMap
	// Error the error happened during the processing of the task.
	// If something goes wrong, the Status() if the Task will also be 'Failed'
	Error error
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
