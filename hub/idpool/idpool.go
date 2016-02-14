package idpool

import (
	"sync"
	"github.com/oleiade/lane"
)

const (
	MAX_QUEUE_LEN = 1000
)


type IdPool interface {

	GetId() uint64
	ReleaseId(id uint64)
}

//pool that gives out incremental ids. Is Thread safe
type IncrementalIdPool struct {
	lastId 		uint64
	lock 		sync.RWMutex
}

/* Thread safe pool that gives incremental IDs, but reuses the dismissed IDs in a FIFO */
type ReusableIdPool struct {
	lastId 		uint64
	freeIds 	*lane.Queue
	lock 		sync.RWMutex
}

//Incremental pool constructor.
func NewIncrementalPool() *IncrementalIdPool{
	return new (IncrementalIdPool)
}

//Reusable pool constructor.
func NewReusableIdPool() *ReusableIdPool{

	pool := &ReusableIdPool{
		lastId: 0, 
		freeIds: lane.NewQueue(), 
	}

	return pool
}

func (pool *IncrementalIdPool) GetId() uint64 {

	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.lastId++ 
	return pool.lastId
}

//Incremental pool does not have implementation for this
func (pool *IncrementalIdPool) ReleaseId(id uint64) {}

func (pool *ReusableIdPool) GetId() uint64 {

	//if available get a free id from the pool
	id := pool.freeIds.Dequeue()

	if id != nil {
		return id.(uint64)
	}

	//if not, return a new sequential id
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.lastId += 1

	return pool.lastId

}

//reuse id only if queue length does not grow excessively
func (pool *ReusableIdPool) ReleaseId(id uint64) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if pool.freeIds.Size() < MAX_QUEUE_LEN && id <= pool.lastId {
		pool.freeIds.Enqueue(id)
	}
}