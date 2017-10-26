package filesystem

import (
	"os"
	"sync"
)

type queue struct {
	mux      sync.Mutex
	monitors map[string]*fileMonitor
}

func (q *queue) addOrUpdate(path string) (bool, error) {
	q.mux.Lock()
	defer q.mux.Unlock()
	m, ok := q.monitors[path]
	if ok {
		m.update()
		return false, nil
	}
	f, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, nil
	}
	q.monitors[path] = newFileMonitor(f, path)
	return f.IsDir(), nil
}

func (q *queue) remove(path string) {
	q.mux.Lock()
	defer q.mux.Unlock()
	delete(q.monitors, path)
}

func (q *queue) all() []*fileMonitor {
	q.mux.Lock()
	defer q.mux.Unlock()
	monitors := make([]*fileMonitor, len(q.monitors))
	i := 0
	for _, m := range q.monitors {
		monitors[i] = m
		i++
	}
	return monitors
}
