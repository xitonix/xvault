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
	encodedFileExtension  = ".xv"
	outputMetadataKey     = "output"
	inputMetadataKey      = "input"
	outputFullMetadataKey = "output_full_path"
	inputFullMetadataKey  = "input_full_path"
)

type FileSet struct {
	Input, Output File
}

type File struct {
	Name, Path string
}

// FilesystemTap is a vault with the functionality of monitoring local filesystem and encrypt the content into the target directory.
// Automatic decryption of the files is not implemented in this vault, because of security reasons.
type FilesystemTap struct {
	pipe           obfuscate.RequestChannel
	master         *obfuscate.MasterKey
	watcher        *watcher.Watcher
	interval       time.Duration
	errors         chan error
	notify         bool
	delete         bool
	source, target string
	wg             *sync.WaitGroup

	openOnce  sync.Once
	closeOnce sync.Once

	//to prevent multiple go routines to run Open and Close at the same time
	mux    sync.Mutex
	isOpen bool
}

// NewFilesystemTap creates a new instance of local storage vault.
// You can feed this vault to a Engine object to automate your encryption tasks.
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
// by the vault if they don't already exist.
func NewFilesystemTap(source, target string,
	pollingInterval time.Duration,
	master *obfuscate.MasterKey,
	notifyErrors bool,
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
		watcher:  w,
		interval: pollingInterval,
		errors:   make(chan error),
		notify:   notifyErrors,
		source:   src,
		target:   tg,
		delete:   deleteCompleted,
		wg:       &sync.WaitGroup{},
		master:   master,
		pipe:     make(obfuscate.RequestChannel),
	}, nil
}

// Errors returns a read-only channel on which you will receive the
// failure notifications. In order to receive the errors on the channel,
// you need to turn error notifications ON by setting
// "notifyErrors" parameter of "NewLocalVault" method to true.
func (f *FilesystemTap) Errors() <-chan error {
	return f.errors
}

func (f *FilesystemTap) Requests() obfuscate.RequestChannel {
	return f.pipe
}

// Open starts the filesystem watcher on the source directory
func (f *FilesystemTap) Open() {
	f.mux.Lock()
	defer f.mux.Unlock()

	if f.isOpen {
		return
	}
	f.openOnce.Do(func() {

		f.wg.Add(1)
		go f.monitorSourceDirectory()

		f.isOpen = true

		// Process the files which are currently in the source folder
		for path, file := range f.watcher.WatchedFiles() {
			f.dispatchWorkUnit(path, file)
		}

		f.wg.Add(1)
		go f.startDirectoryWatcher()
	})
}

// Close stops the filesystem watcher and releases the resources.
// NOTE: You don't need to explicitly call this function when you are using this vault
// with a "Engine". The processor will take care of it
func (f *FilesystemTap) Close() {
	f.mux.Lock()
	defer f.mux.Unlock()

	if !f.isOpen {
		return
	}
	f.closeOnce.Do(func() {
		if f != nil && f.watcher != nil {
			f.isOpen = false
			f.watcher.Close()
			f.wg.Wait()
			close(f.pipe)
			close(f.errors)
		}
	})
}

func (f *FilesystemTap) IsOpen() bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.isOpen
}

func (f *FilesystemTap) startDirectoryWatcher() {
	defer f.wg.Done()
	err := f.watcher.Start(f.interval)

	if err != nil {
		f.reportError(fmt.Errorf("filesystem watcher: %s", err))
	}
}

func (f *FilesystemTap) monitorSourceDirectory() {
	defer f.wg.Done()
	for {
		select {
		case event := <-f.watcher.Event:
			f.dispatchWorkUnit(event.Path, event.FileInfo)
		case err := <-f.watcher.Error:
			f.reportError(err)
		case <-f.watcher.Closed:
			return
		}
	}
}

func (f *FilesystemTap) reportError(err error) {
	if f.isOpen && f.notify {
		f.errors <- err
	}
}

// whenDone is a callback method which will get called by the processor once the
// processing of a task has been finished
func (f *FilesystemTap) whenDone(w *obfuscate.WorkUnit) {
	m := f.ParseMetadata(w.Metadata)

	err := w.Task.CloseInput()
	if err != nil {
		f.reportError(fmt.Errorf("failed to close '%s': %s", m.Input.Name, err))
	}
	err = w.Task.CloseOutputs()
	if err != nil {
		f.reportError(fmt.Errorf("failed to close '%v': %s", m.Output.Name, err))
	}

	if f.delete && w.Task.Status() == obfuscate.Completed {
		file := m.Input.Path
		err := os.Remove(file)
		if err != nil {
			f.reportError(fmt.Errorf("failed to remove '%s': %s", m.Input.Name, err))
		}
		dir := filepath.Dir(file)
		isEmpty := isDirEmpty(dir)
		if isEmpty && dir != f.source {
			os.RemoveAll(dir)
		}
	}
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

	if !f.isOpen || (f.source == path) || file.IsDir() {
		return
	}

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
	w.Metadata[inputMetadataKey] = name
	w.Metadata[outputMetadataKey] = name + encodedFileExtension
	w.Metadata[inputFullMetadataKey] = inputFullPath
	w.Metadata[outputFullMetadataKey] = outputFullPath
	f.pipe <- w
}

func (f *FilesystemTap) ParseMetadata(metadata obfuscate.MetadataMap) FileSet {
	return FileSet{
		Input: File{
			Name: metadata[inputMetadataKey].(string),
			Path: metadata[inputFullMetadataKey].(string),
		},
		Output: File{
			Name: metadata[outputMetadataKey].(string),
			Path: metadata[outputFullMetadataKey].(string),
		},
	}
}

func (f *FilesystemTap) createTargetSubDirectory(path, name string) {
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
