package syncmap

import(
	"sync"
)

/* Provide a simple thread safe map for generic types*/
type SyncMap struct {
	lock sync.RWMutex
	m    map[uint64]interface{}
}

func NewSyncMap() *SyncMap {
	return &SyncMap{m: make(map[uint64]interface{})}
}

func (s *SyncMap) Get(key uint64) (interface{}, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok := s.m[key]
	return value, ok
}

func (s *SyncMap) Set(key uint64, value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m[key] = value
}

func (s *SyncMap) Remove(key uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.m, key)
}

func (s *SyncMap) Length() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	
	return len(s.m)
}

func (s *SyncMap) GetKeys() []uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	
	keys := make([]uint64, len(s.m))
	var i int
	for k := range s.m {
		keys[i] = k
		i++
	}

	return keys

}