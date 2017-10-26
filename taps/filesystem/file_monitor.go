package filesystem

import (
	"os"
	"sync"
	"time"
)

type fileMonitor struct {
	path       string
	fi         os.FileInfo
	mux        sync.Mutex
	lastUpdate time.Time
}

func newFileMonitor(fi os.FileInfo, path string) *fileMonitor {
	return &fileMonitor{
		path:       path,
		fi:         fi,
		lastUpdate: time.Now(),
	}
}

func (m *fileMonitor) update() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.lastUpdate = time.Now()
}

func (m *fileMonitor) isReady() bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	return !m.lastUpdate.IsZero() && time.Now().Sub(m.lastUpdate) > (2*time.Second)
}
