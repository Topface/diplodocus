package diplodocus

import (
	"github.com/fsnotify/fsnotify"
	"sync"
)

// FileManager is an entry point for module, it manipulates with files.
type FileManager struct {
	watcher    *fsnotify.Watcher
	responders *ResponderMap
	files      map[string]*file
	mutex      sync.Mutex
}

// NewFileManager creates file manager with specified fs watcher.
func NewFileManager(watcher *fsnotify.Watcher) *FileManager {
	return &FileManager{
		watcher:    watcher,
		responders: newResponderMap(),
		files:      map[string]*file{},
	}
}

// Watch monitors and processes fs events before it encounters an error.
func (m *FileManager) Watch() error {
	for {
		select {
		case ev := <-m.watcher.Events:
			files := m.responders.GetMappings(ev.Name)
			for _, file := range files {
				file.OnEvent(ev)
			}
		case err := <-m.watcher.Errors:
			return err
		}
	}
}

// GetFile returns File object for specified fs path.
func (m *FileManager) GetFile(path string) (*file, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if f, ok := m.files[path]; ok {
		return f, nil
	}

	f, err := newFile(path, m.responders)
	if err != nil {
		return nil, err
	}

	m.files[path] = f

	return f, nil
}
