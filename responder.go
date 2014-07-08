package diplodocus

import "sync"

// ResponderMap keeps track of file objects that are responsible
// for processing of paths on filesystem. Only useful if logs
// are rotated with symlink shifting, like in scribe.
type ResponderMap struct {
	mapping map[string][]*file
	mutex   sync.Mutex
}

// newResponderMap returns empty responder map.
func newResponderMap() *ResponderMap {
	return &ResponderMap{
		mapping: make(map[string][]*file),
	}
}

// AddMapping adds mapping between path on filesystem and file object.
// Should be called from file object on opening.
func (r *ResponderMap) AddMapping(path string, f *file) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if files, ok := r.mapping[path]; ok {
		files = append(files, f)
		r.mapping[path] = files
	} else {
		r.mapping[path] = []*file{f}
	}
}

// GetMappings returns file objects responsible for specified path.
func (r *ResponderMap) GetMappings(path string) []*file {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if files, ok := r.mapping[path]; ok {
		return files
	} else {
		return []*file{}
	}
}

// RemoveMapping removes association between path on filesystem
// and file object that is responsible for processing.
func (r *ResponderMap) RemoveMapping(path string, f *file) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if files, ok := r.mapping[path]; ok {
		for i, current := range files {
			if current == f {
				r.mapping[path] = append(files[:i], files[i+1:]...)
			}
		}
	}
}
