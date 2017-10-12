package taps

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/xitonix/xvault/obfuscate"
	"io/ioutil"
	"strings"
)

const (
	encodedFileExtension = ".xv"
	outputMetadataKey    = "output"
	inputMetadataKey     = "input"
)

// Report represents the progress details of a task
type Report struct {
	// Status the status of the operation
	Status obfuscate.Status
	// Input the name of the input file
	Input string
	// Output the name of the output file
	Output string
	// Error the error details of a failed task
	Error error
}

// FilesystemTap is a vault with the functionality of monitoring local filesystem and encrypt the content into the target directory.
// Automatic decryption of the files is not implemented in this vault, because of security reasons.
type FilesystemTap struct {
	pipe           *obfuscate.Pipe
	progress       chan Report
	watcher        *watcher.Watcher
	interval       time.Duration
	errors         chan error
	notify, report bool
	delete         bool
	source, target string
	wg             *sync.WaitGroup

	mux    sync.Mutex
	isOpen bool
}

// NewFilesystemTap creates a new instance of local storage vault.
// You can feed this vault to a WorkBucket object to automate your encryption tasks.
//
// If you have enabled error notification by setting  'notifyErrors' to true, you need to make sure
// that you subscribe to "Errors" channel to read off the notification pipe, otherwise you will get
// get blocked on the full channel. Same rule applies to "reportProgress" flag. You will need to subscribe to
// "Progress" channel to avoid blocking the vault operations, if you set the flag to true.
//
// "parallelism" specifies the number of files you would like to process at the same time.
//
// "pollingInterval" is the frequency of checking the "source" directory for newly created files.
//
// if you set "deleteCompleted" to true, the input files will get deleted, only if the encryption
// operation has been finished successfully.
//
// "source" and "target" are the paths to source and destination directories. They will get created
// by the vault if they don't already exist.
func NewFilesystemTap(source, target string,
	pollingInterval time.Duration,
	notifyErrors bool,
	reportProgress bool,
	deleteCompleted bool) (*FilesystemTap, error) {
	src, err := createDirIfNotExist(source)
	if err != nil {
		return nil, err
	}

	tg, err := createDirIfNotExist(target)
	if err != nil {
		return nil, err
	}

	w := watcher.New()
	w.FilterOps(watcher.Create)
	w.IgnoreHiddenFiles(true)

	if err := w.AddRecursive(src); err != nil {
		return nil, err
	}

	return &FilesystemTap{
		progress: make(chan Report),
		watcher:  w,
		interval: pollingInterval,
		errors:   make(chan error),
		notify:   notifyErrors,
		report:   reportProgress,
		source:   src,
		target:   tg,
		delete:   deleteCompleted,
		wg:       &sync.WaitGroup{},
	}, nil
}

// Errors returns a read-only channel on which you will receive the
// failure notifications. In order to receive the errors on the channel,
// you need to turn error notifications ON by setting
// "notifyErrors" parameter of "NewLocalVault" method to true.
func (f *FilesystemTap) Errors() <-chan error {
	return f.errors
}

// Open starts the filesystem watcher on the source directory
func (f *FilesystemTap) Open(pipe *obfuscate.Pipe) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.isOpen {
		return
	}
	f.pipe = pipe
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			select {
			case event := <-f.watcher.Event:
				f.dispatchWorkUnit(event.Path, event.FileInfo)
			case err := <-f.watcher.Error:
				if f.notify {
					f.errors <- err
				}
			case <-f.watcher.Closed:
				return
			}
		}
	}()

	// Process the files which are currently in the source folder
	for path, file := range f.watcher.WatchedFiles() {
		f.dispatchWorkUnit(path, file)
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		err := f.watcher.Start(f.interval)

		if err != nil && f.notify {
			f.errors <- fmt.Errorf("filesystem watcher: %s", err)
		}
	}()
	f.isOpen = true
}

// Close stops the filesystem watcher and releases the resources.
// NOTE: You don't need to explicitly call this function when you are using this vault
// with a "WorkBucket". The processor will take care of it
func (f *FilesystemTap) Close() {
	f.mux.Lock()
	defer f.mux.Unlock()
	if !f.isOpen {
		return
	}
	if f != nil && f.watcher != nil {
		f.watcher.Close()
		f.wg.Wait()
		close(f.errors)
		close(f.progress)
	}
	f.isOpen = false
}

func (f *FilesystemTap) IsOpen() bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.isOpen
}

// Progress returns a channel on which you will receive the result of processing each task.
// The failed tasks will be dispatched on this channel as well.
// NOTE: In order to receive the progress report on this channel,
// you need to turn reporting ON by setting
// "reportProgress" parameter of "NewLocalVault" method to true.
func (f *FilesystemTap) Progress() <-chan Report {
	return f.progress
}

// whenDone is a callback method which will get called by the processor once the
// processing of a task has been finished
func (f *FilesystemTap) whenDone(task *obfuscate.Task) {
	input := task.Name()
	output := task.MetaData[outputMetadataKey].(string)

	err := task.CloseInput()
	if err != nil && f.notify {
		f.errors <- fmt.Errorf("failed to close '%s': %s", input, err)
	}
	err = task.CloseOutputs()
	if err != nil && f.notify {
		f.errors <- fmt.Errorf("failed to close '%v': %s", output, err)
	}

	if f.delete && task.Status() == obfuscate.Completed {
		file := task.MetaData[inputMetadataKey].(string)
		err := os.Remove(file)
		if err != nil && f.notify {
			f.errors <- fmt.Errorf("failed to remove '%s': %s", input, err)
		}
		dir := filepath.Dir(file)
		isEmpty := isDirEmpty(dir)
		if isEmpty && dir != f.source {
			os.RemoveAll(dir)
		}
	}

	f.reportProgress(task)
}

func (f *FilesystemTap) openInputFile(path string) (*os.File, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, path, err
	}

	//We need to wait until the watcher releases the handle
	time.Sleep(100 * time.Millisecond)

	input, err := os.Open(abs)
	return input, abs, err
}

func (f *FilesystemTap) createOutputFile(name, inputFullPath string) (*os.File, string, error) {
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

func (f *FilesystemTap) dispatchWorkUnit(path string, file os.FileInfo) {

	if (f.source == path) || file.IsDir() {
		return
	}

	input, inputFullPath, err := f.openInputFile(path)
	if err != nil && f.notify {
		f.errors <- fmt.Errorf("failed to open '%s': %s", path, err)
		return
	}

	name := file.Name()

	output, outputFullPath, err := f.createOutputFile(name, inputFullPath)
	if err != nil && f.notify {
		f.errors <- fmt.Errorf("failed to create '%s': %s", outputFullPath, err)
		return
	}

	t := obfuscate.NewTask(name, obfuscate.Encode, input, output)
	t.AddMetadata(outputMetadataKey, name+encodedFileExtension)
	t.AddMetadata(inputMetadataKey, inputFullPath)
	f.reportProgress(t)
	f.pipe.Push(obfuscate.NewWorkUnit(t, f.whenDone))
}

func (f *FilesystemTap) createTargetSubDirectory(path, name string) {
	abs, err := filepath.Abs(filepath.Join(f.target, name))
	if err != nil && f.notify {
		f.errors <- fmt.Errorf("failed to resolve the path to '%s': %s", path, err)
		return
	}

	dir, err := createDirIfNotExist(abs)
	if err != nil && f.notify {
		f.errors <- fmt.Errorf("failed to create '%s': %s", dir, err)
		return
	}
}

func (f *FilesystemTap) reportProgress(t *obfuscate.Task) {
	if f.report {
		f.progress <- Report{
			Status: t.Status(),
			Error:  t.Error,
			Input:  t.Name(),
			Output: t.MetaData[outputMetadataKey].(string),
		}
	}
}

func createDirIfNotExist(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir, err
	}
	f, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return abs, os.MkdirAll(abs, os.ModePerm)
	}
	if !f.IsDir() {
		return abs, ErrInvalidDirectory
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
