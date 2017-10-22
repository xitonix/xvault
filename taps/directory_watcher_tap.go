package taps

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/rjeczalik/notify"
	"github.com/xitonix/xvault/obfuscate"
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

// DirectoryWatcherTap is a tap with the functionality of monitoring local filesystem and encrypting the content into the target directory.
type DirectoryWatcherTap struct {
	pipe           obfuscate.WorkList
	progress       chan *Result
	master         *obfuscate.MasterKey
	interval       time.Duration
	errors         chan error
	notifyErr      bool
	report         bool
	delete         bool
	source, target string
	wg             *sync.WaitGroup
	done           chan obfuscate.None
	fsEvents       chan notify.EventInfo

	openOnce  sync.Once
	closeOnce sync.Once

	// to prevent multiple go routines to run
	// Open and Close at the same time
	mux    sync.Mutex
	isOpen bool
}

// NewDirectoryWatcherTap creates a new instance of directory watcher tap.
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
func NewDirectoryWatcherTap(source, target string,
	pollingInterval time.Duration,
	master *obfuscate.MasterKey,
	notifyErrors bool,
	reportProgress bool,
	deleteCompleted bool) (*DirectoryWatcherTap, error) {
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

	return &DirectoryWatcherTap{
		interval:  pollingInterval,
		errors:    make(chan error),
		notifyErr: notifyErrors,
		source:    src,
		target:    tg,
		delete:    deleteCompleted,
		wg:        &sync.WaitGroup{},
		master:    master,
		pipe:      make(obfuscate.WorkList),
		progress:  make(chan *Result),
		done:      make(chan obfuscate.None),
		// Make the channel buffered to ensure no event is dropped. Notify will drop
		// an event if the receiver is not able to keep up the sending pace.
		fsEvents: make(chan notify.EventInfo, 1),

		report: reportProgress,
	}, nil
}

// Errors returns a read-only channel on which you will receive the failure notifications.
//
// In order to receive the errors on the channel, you need to turn error notifications On by setting
// "notifyErrors" parameter of "NewDirectoryWatcherTap" method to true.
// You can also switch it On or Off by calling the SwitchErrorNotification(...) method
func (d *DirectoryWatcherTap) Errors() <-chan error {
	return d.errors
}

// Pipe returns the work list channel from which the engine will receive the encryption requests.
func (d *DirectoryWatcherTap) Pipe() obfuscate.WorkList {
	return d.pipe
}

// Progress returns a read-only channel on which you will receive the progress report
//
// In order to receive progress report on the channel, you need to turn it On by setting
// "reportProgress" parameter of "NewDirectoryWatcherTap" method to true.
// You can also switch it On or Off by calling the SwitchProgressReport(...) method
func (d *DirectoryWatcherTap) Progress() <-chan *Result {
	return d.progress
}

// SwitchErrorNotification switches error notification ON/OFF
func (d *DirectoryWatcherTap) SwitchErrorNotification(on bool) {
	d.notifyErr = on
}

// SwitchProgressReport switches progress report ON/OFF
func (d *DirectoryWatcherTap) SwitchProgressReport(on bool) {
	d.report = on
}

// Open starts the directory watcher on the source directory.
// You SHOULD NOT call this method explicitly when you use the tap with an Engine object.
// Starting the engine will take care of opening the tap.
func (d *DirectoryWatcherTap) Open() {
	d.mux.Lock()
	defer d.mux.Unlock()

	d.openOnce.Do(func() {

		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			// Process the files which are currently in the source folder
			d.processExistingFiles()
		}()

		d.wg.Add(1)
		go d.startDirectoryWatcher()

		d.isOpen = true
	})
}
func (d *DirectoryWatcherTap) processExistingFiles() {
	// TODO Walk src
}

// Close stops the filesystem watcher and releases the resources.
// NOTE: You don't need to explicitly call this function when you are using the tap
// with an Engine
func (d *DirectoryWatcherTap) Close() {
	d.mux.Lock()
	defer d.mux.Unlock()

	d.closeOnce.Do(func() {
		if d != nil {
			notify.Stop(d.fsEvents)
			close(d.done)
			d.wg.Wait()
			close(d.pipe)
			close(d.errors)
			close(d.progress)
			d.isOpen = false
		}
	})
}

// IsOpen returns true if the tap is open
func (d *DirectoryWatcherTap) IsOpen() bool {
	d.mux.Lock()
	defer d.mux.Unlock()
	return d.isOpen
}

func (d *DirectoryWatcherTap) reportProgress(r *Result) {
	d.progress <- r
}

func (d *DirectoryWatcherTap) startDirectoryWatcher() {
	defer d.wg.Done()

	if err := notify.Watch("src/...", d.fsEvents, notify.Create, notify.Write); err != nil {
		d.reportError(err)
		d.Close()
		return
	}
	files := make(map[string]time.Time)
	ticker := time.NewTicker(1 * time.Second)
	var processing bool
	defer ticker.Stop()
	for {
		select {
		case <-d.done:
			return
		case ei := <-d.fsEvents:
			files[ei.Path()] = time.Now()
		case <-ticker.C:
			if processing {
				continue
			}
			processing = true
			now := time.Now()
			if len(files) > 0 {
				fmt.Println(len(files))
			}
			for f, t := range files {
				if now.Sub(t) >= time.Second {
					fi, err := os.Stat(f)
					if err != nil {
						d.reportError(fmt.Errorf("os.Stat: %s", err))
						continue
					}
					if !fi.IsDir() {
						fmt.Println("**", f)
						d.dispatchWorkUnit(f, fi)
					}
					delete(files, f)
				}
			}
			processing = false
		}
	}
}

func (d *DirectoryWatcherTap) reportError(err error) {
	if d.IsOpen() && d.notifyErr {
		d.errors <- err
	}
}

// whenDone is a callback method which will get called by the processor once the
// processing of a task has been finished
func (d *DirectoryWatcherTap) whenDone(w *obfuscate.WorkUnit) {
	input, output := d.parseMetadata(w.Metadata)

	err := w.Task.CloseInput()
	if err != nil {
		d.reportError(fmt.Errorf("failed to close '%s': %s", input.Name, err))
	}
	err = w.Task.CloseOutputs()
	if err != nil {
		d.reportError(fmt.Errorf("failed to close '%v': %s", output.Name, err))
	}

	if d.delete && w.Task.Status() == obfuscate.Completed {
		file := input.Path
		err := os.Remove(file)
		if err != nil {
			d.reportError(fmt.Errorf("failed to remove '%s': %s", input.Name, err))
		}
		//d.removeDirectory(file)
	}

	if d.report && d.IsOpen() {
		d.reportProgress(&Result{
			Output: output,
			Input:  input,
			Status: w.Task.Status(),
			Error:  w.Error,
		})
	}
}
func (d *DirectoryWatcherTap) removeDirectory(file string) {
	dir := filepath.Dir(file)
	if dir == d.source || !isDirEmpty(dir) {
		return
	}

	err := os.RemoveAll(dir)
	if err != nil {
		d.reportError(fmt.Errorf("failed to remove '%s' directory: %s", dir, err))
	}
	d.removeDirectory(dir)
}

func (d *DirectoryWatcherTap) openInputFile(path string) (*os.File, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, path, err
	}

	input, err := os.Open(abs)
	return input, abs, err
}

func (d *DirectoryWatcherTap) createOutputFile(name, inputFullPath string) (*os.File, string, error) {
	subDir := strings.Replace(filepath.Dir(inputFullPath), d.source, "", 1)
	targetDir := filepath.Join(d.target, subDir)
	abs, err := createDirIfNotExist(targetDir)
	if err != nil {
		return nil, name, err
	}
	abs = filepath.Join(abs, name+encodedFileExtension)
	output, err := os.Create(abs)
	return output, abs, err
}

func (d *DirectoryWatcherTap) dispatchWorkUnit(path string, file os.FileInfo) {

	input, inputFullPath, err := d.openInputFile(path)
	if err != nil {
		d.reportError(fmt.Errorf("failed to open '%s': %s", path, err))
		return
	}

	name := file.Name()

	output, outputFullPath, err := d.createOutputFile(name, inputFullPath)
	if err != nil {
		d.reportError(fmt.Errorf("failed to create '%s': %s", outputFullPath, err))
		return
	}

	t := obfuscate.NewTask(obfuscate.Encode, input, output)
	w := obfuscate.NewWorkUnit(t, d.master, d.whenDone)
	outName := name + encodedFileExtension
	w.Metadata[inputMetadataKey] = name
	w.Metadata[outputMetadataKey] = outName
	w.Metadata[inputFullMetadataKey] = inputFullPath
	w.Metadata[outputFullMetadataKey] = outputFullPath
	if d.report {
		d.reportProgress(&Result{
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
	}

	d.pipe <- w
}

func (d *DirectoryWatcherTap) parseMetadata(metadata obfuscate.MetadataMap) (File, File) {
	return File{
			Name: metadata[inputMetadataKey].(string),
			Path: metadata[inputFullMetadataKey].(string),
		},
		File{
			Name: metadata[outputMetadataKey].(string),
			Path: metadata[outputFullMetadataKey].(string),
		}
}

func (d *DirectoryWatcherTap) createTargetSubDirectory(path, name string) {
	abs, err := filepath.Abs(filepath.Join(d.target, name))
	if err != nil {
		d.reportError(fmt.Errorf("failed to resolve the path to '%s': %s", path, err))
		return
	}

	dir, err := createDirIfNotExist(abs)
	if err != nil {
		d.reportError(fmt.Errorf("failed to create '%s': %s", dir, err))
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
	return strings.TrimRight(abs, `/\`), err
}

func isDirEmpty(name string) bool {
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return false
	}
	return len(entries) == 0
}
