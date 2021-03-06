package diplodocus

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"os"
	"path"
	"sync"
	"time"
)

// file is a structure to follow append-only files on file system.
// It sends events with new data chunks to listeners and handles
// data appends, file removals and symlink changes.
type file struct {
	path       string
	real       string
	file       *os.File
	size       int64
	listeners  []Listener
	responders *ResponderMap
	mapped     bool
	mutex      sync.Mutex
	events     chan event
}

// newFile creates file for specified path and with specified responder.
func newFile(path string, responders *ResponderMap) (*file, error) {
	f := &file{
		path:       path,
		responders: responders,
		events:     make(chan event),
	}

	go f.readPipe()

	err := f.open()
	if err != nil {
		return nil, err
	}

	return f, nil
}

// open opens file, gets its size, and handles symlink.
func (f *file) open() error {
	file, err := os.Open(f.path)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	info, err := os.Lstat(f.path)
	if err != nil {
		return err
	}

	real := f.path
	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		real, err = os.Readlink(f.path)
		if err != nil {
			return err
		}

		if !path.IsAbs(real) {
			real = path.Join(path.Dir(f.path), real)
		}
	}

	f.file = file
	f.real = real
	f.size = stat.Size()

	if !f.mapped {
		f.mapped = true
		f.responders.AddMapping(f.path, f)
	}

	if f.path != f.real {
		f.responders.AddMapping(f.real, f)
	}

	return nil
}

// AddListener adds a listener for file events.
func (f *file) AddListener(listener Listener) {
	f.mutex.Lock()
	f.listeners = append(f.listeners, listener)
	f.mutex.Unlock()
}

// RemoveListener removes listener for file events.
func (f *file) RemoveListener(listener Listener) {
	f.mutex.Lock()

	for i, current := range f.listeners {
		if listener == current {
			f.listeners = append(f.listeners[:i], f.listeners[i+1:]...)
		}
	}

	f.mutex.Unlock()
}

// OnEvent should be called when event associated with this file happens.
func (f *file) OnEvent(ev fsnotify.Event) {
	f.mutex.Lock()

	if ev.Op&fsnotify.Remove == fsnotify.Remove || ev.Op&fsnotify.Rename == fsnotify.Rename {
		f.file = nil
		f.size = 0
		if f.path != f.real {
			f.responders.RemoveMapping(f.real, f)
		}

		f.mutex.Unlock()

		// file removed
		return
	}

	if f.file == nil {
		err := f.open()
		if err != nil {
			f.mutex.Unlock()
			f.events <- event{Error: errors.New("file open failed")}
			return
		}
	}

	if ev.Op&fsnotify.Create == fsnotify.Create {
		f.size = 0
	}

	stat, err := f.file.Stat()
	if err != nil {
		f.mutex.Unlock()
		f.events <- event{Error: err}
		return
	}

	off := f.size
	f.size = stat.Size()

	if len(f.listeners) == 0 {
		f.mutex.Unlock()
		return
	}

	f.mutex.Unlock()

	if f.size == off {
		return
	}

	if f.size <= off {
		// file truncated
		return
	}

	buf := make([]byte, f.size-off)
	_, err = f.file.ReadAt(buf, off)
	if err != nil {
		f.events <- event{Error: err}
		return
	}

	f.events <- event{Buffer: &buf}
}

// readPipe reads events from channel and distributes them among listeners.
func (f *file) readPipe() {
	for ev := range f.events {
		f.mutex.Lock()
		listeners := make([]Listener, len(f.listeners))
		copy(listeners, f.listeners)
		f.mutex.Unlock()

		wg := sync.WaitGroup{}

		for _, listener := range listeners {
			wg.Add(1)

			go func(listener Listener) {
				select {
				case listener <- ev:
				case <-time.After(time.Second * 10):
					f.RemoveListener(listener)
					close(listener)
				}

				wg.Done()
			}(listener)
		}

		wg.Wait()
	}
}
