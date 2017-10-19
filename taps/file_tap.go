package taps

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xitonix/xvault/obfuscate"
	"sync/atomic"
)

const (
	encodedFileExtension  = ".xv"
	outputMetadataKey     = "output"
	inputMetadataKey      = "input"
	outputFullMetadataKey = "output_full_path"
	inputFullMetadataKey  = "input_full_path"
)

type File struct {
	Name, Path string
}

// Result represents the progress details of a task
type Result struct {
	// Status the status of the operation
	Status obfuscate.Status

	// Error the error details of a failed task
	Error error

	Input, Output File
}

// FileTap is a tap with the functionality of monitoring local filesystem and encrypting the content into the target directory.
type FileTap struct {
	pipe            obfuscate.WorkList
	progress        chan *Result
	master          *obfuscate.MasterKey
	interval        time.Duration
	errors          chan error
	notifyErr       bool
	report          bool
	delete          bool
	source, target  string
	wg              *sync.WaitGroup
	pollingInterval time.Duration
	progressCounter int32
	clean           chan obfuscate.None

	openOnce  sync.Once
	closeOnce sync.Once

	// to prevent multiple go routines to run
	// Open and Close at the same time
	mux    sync.Mutex
	isOpen bool
}

// NewFileTap creates a new instance of directory watcher tap.
//
// If you have enabled error notification by setting  'notifyErrors' to true, you need to make sure
// that you subscribe to "Errors" channel to read off the notification pipe, otherwise you will get
// get blocked on the full channel.
//
// "parallelism" specifies the number of files you would like to process at the same time.
//
// "pollingInterval" is the frequency of checking the "source" directory for newly created files.
//
// if you set "deleteCompleted" to true, the input files will get deleted, only if the encryption
// operation has been finished successfully.
//
// "source" and "target" are the paths to source and destination directories. They will get created
// by the tap if they don't already exist.
func NewFileTap(source, target string,
	pollingInterval time.Duration,
	master *obfuscate.MasterKey,
	notifyErrors bool,
	reportProgress bool,
	deleteCompleted bool) (*FileTap, error) {
	src, err := createDirIfNotExist(source)
	if err != nil {
		return nil, err
	}

	tg, err := createDirIfNotExist(target)
	if err != nil {
		return nil, err
	}

	return &FileTap{
		interval:        pollingInterval,
		errors:          make(chan error),
		notifyErr:       notifyErrors,
		source:          src,
		target:          tg,
		delete:          deleteCompleted,
		wg:              &sync.WaitGroup{},
		master:          master,
		pipe:            make(obfuscate.WorkList),
		progress:        make(chan *Result),
		report:          reportProgress,
		clean:           make(chan obfuscate.None),
		pollingInterval: pollingInterval,
	}, nil
}

// Errors returns a read-only channel on which you will receive the failure notifications.
//
// In order to receive the errors on the channel, you need to turn error notifications On by setting
// "notifyErrors" parameter of "NewFileTap" method to true.
// You can also switch it On or Off by calling the SwitchErrorNotification(...) method
func (f *FileTap) Errors() <-chan error {
	return f.errors
}

// Pipe returns the work list channel from which the engine will receive the encryption requests.
func (f *FileTap) Pipe() obfuscate.WorkList {
	return f.pipe
}

// Progress returns a read-only channel on which you will receive the progress report
//
// In order to receive progress report on the channel, you need to turn it On by setting
// "reportProgress" parameter of "NewFileTap" method to true.
// You can also switch it On or Off by calling the SwitchProgressReport(...) method
func (f *FileTap) Progress() <-chan *Result {
	return f.progress
}

// SwitchErrorNotification switches error notification ON/OFF
func (f *FileTap) SwitchErrorNotification(on bool) {
	f.notifyErr = on
}

// SwitchProgressReport switches progress report ON/OFF
func (f *FileTap) SwitchProgressReport(on bool) {
	f.report = on
}

// Open starts the directory watcher on the source directory.
// You SHOULD NOT call this method explicitly when you use the tap with an Engine object.
// Starting the engine will take care of opening the tap.
func (f *FileTap) Open() {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.openOnce.Do(func() {

		f.wg.Add(1)
		go f.startTheCleaner()

		f.isOpen = true
	})
}

// Close stops the filesystem watcher and releases the resources.
// NOTE: You don't need to explicitly call this function when you are using the tap
// with an Engine
func (f *FileTap) Close() {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.closeOnce.Do(func() {
		if f != nil {
			f.isOpen = false
			close(f.clean)
			f.wg.Wait()
			close(f.pipe)
			close(f.errors)
			close(f.progress)
		}
	})
}

// IsOpen returns true if the tap is open
func (f *FileTap) IsOpen() bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.isOpen
}

func (f *FileTap) Process() {
	f.wg.Add(1)
	defer f.wg.Done()
	err := filepath.Walk(f.source, func(path string, info os.FileInfo, err error) error {

		if !f.isOpen  {
			return io.EOF
		}
		if err != nil {
			f.reportError(err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		f.dispatchWorkUnit(path, info)
		return nil
	})

	if err != nil {
		if err != io.EOF {
			f.reportError(err)
		}
		return
	}

	if f.delete && f.isOpen {
		f.clean <- obfuscate.None{}
	}
}

func (f *FileTap) reportProgress(r *Result) {
	if f.isOpen && f.report {
		f.progress <- r
	}
}

func (f *FileTap) reportError(err error) {
	if f.IsOpen() && f.notifyErr {
		f.errors <- err
	}
}

// whenDone is a callback method which will get called by the processor once the
// processing of a task has been finished
func (f *FileTap) whenDone(w *obfuscate.WorkUnit) {
	input, output := f.parseMetadata(w.Metadata)

	err := w.Task.CloseInput()
	if err != nil {
		f.reportError(fmt.Errorf("failed to close '%s': %s", input.Name, err))
	}
	err = w.Task.CloseOutputs()
	if err != nil {
		f.reportError(fmt.Errorf("failed to close '%v': %s", output.Name, err))
	}

	if f.delete && w.Task.Status() == obfuscate.Completed {
		file := input.Path
		err := os.Remove(file)
		if err != nil {
			f.reportError(fmt.Errorf("failed to remove '%s': %s", input.Name, err))
		} else {
			f.oneDone()
		}
	}

	f.reportProgress(&Result{
		Output: output,
		Input:  input,
		Status: w.Task.Status(),
		Error:  w.Error,
	})
}

func (f *FileTap) oneDone() {
	atomic.AddInt32(&f.progressCounter, -1)
}

func (f *FileTap) oneStarted() {
	atomic.AddInt32(&f.progressCounter, 1)
}

func (f *FileTap) resetInProgressFlag() {
	atomic.StoreInt32(&f.progressCounter, 0)
}

func (f *FileTap) isInProgress() bool {
	return atomic.LoadInt32(&f.progressCounter) == 0
}

func (f *FileTap) openInputFile(path string) (*os.File, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, path, err
	}

	time.Sleep(100 * time.Millisecond)

	input, err := os.Open(abs)
	return input, abs, err
}

func (f *FileTap) createOutputFile(name, inputFullPath string) (*os.File, string, error) {
	subDir := strings.Replace(filepath.Dir(inputFullPath), f.source, "", 1)
	targetDir := filepath.Join(f.target, subDir)
	abs, err := createDirIfNotExist(targetDir)
	if err != nil {
		return nil, name, err
	}
	abs = filepath.Join(abs, name+encodedFileExtension)
	output, err := os.Create(abs)
	return output, abs, err
}

func (f *FileTap) dispatchWorkUnit(path string, file os.FileInfo) {

	input, inputFullPath, err := f.openInputFile(path)
	if err != nil {
		f.reportError(fmt.Errorf("failed to open '%s': %s", path, err))
		return
	}

	name := file.Name()

	output, outputFullPath, err := f.createOutputFile(name, inputFullPath)
	if err != nil {
		f.reportError(fmt.Errorf("failed to create '%s': %s", outputFullPath, err))
		return
	}

	t := obfuscate.NewTask(obfuscate.Encode, input, output)
	w := obfuscate.NewWorkUnit(t, f.master, f.whenDone)
	outName := name + encodedFileExtension
	w.Metadata[inputMetadataKey] = name
	w.Metadata[outputMetadataKey] = outName
	w.Metadata[inputFullMetadataKey] = inputFullPath
	w.Metadata[outputFullMetadataKey] = outputFullPath
	f.reportProgress(&Result{
		Status: t.Status(),
		Input: File{
			Name: name,
			Path: inputFullPath,
		},
		Output: File{
			Name: outName,
			Path: outputFullPath,
		},
	})

	f.oneStarted()
	f.pipe <- w
}

func (f *FileTap) parseMetadata(metadata obfuscate.MetadataMap) (File, File) {
	return File{
			Name: metadata[inputMetadataKey].(string),
			Path: metadata[inputFullMetadataKey].(string),
		},
		File{
			Name: metadata[outputMetadataKey].(string),
			Path: metadata[outputFullMetadataKey].(string),
		}
}

func (f *FileTap) createTargetSubDirectory(path, name string) {
	abs, err := filepath.Abs(filepath.Join(f.target, name))
	if err != nil {
		f.reportError(fmt.Errorf("failed to resolve the path to '%s': %s", path, err))
		return
	}

	dir, err := createDirIfNotExist(abs)
	if err != nil {
		f.reportError(fmt.Errorf("failed to create '%s': %s", dir, err))
		return
	}
}

func createDirIfNotExist(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir, err
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return abs, os.MkdirAll(abs, os.ModePerm)
	}
	return abs, err
}

func isDirEmpty(name string) bool {
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

func (f *FileTap) startTheCleaner() {
	defer f.wg.Done()
	for {
		select {
		case _, more := <-f.clean:
			if !more {
				return
			}
			f.cleanup()
		}
	}
}

func (f *FileTap) cleanup() {
	var tried int
	const times = 3
	for tried < times && f.isInProgress() {
		select {
		case <-time.After(10 * time.Second):
			if !f.isInProgress() {
				tried = times + 3
			} else {
				tried++
			}
		case _, more := <-f.clean:
			if !more {
				return
			}
		}
	}

	if !f.isOpen {
		return
	}

	err := filepath.Walk(f.source, func(path string, info os.FileInfo, err error) error {
		if !f.isOpen {
			return io.EOF
		}

		if info.IsDir() && path != f.source {
			if isDirEmpty(path) && f.isOpen {
				err := os.RemoveAll(path)
				if err != nil && !os.IsNotExist(err) {
					f.reportError(err)
				}
			}
		}
		return nil
	})

	if err != nil && err != io.EOF {
		f.reportError(err)
	}
}
