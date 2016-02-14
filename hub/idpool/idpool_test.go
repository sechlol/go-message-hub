package idpool_test

import(
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/hub/idpool"
)

var p1 = idpool.NewIncrementalPool()
var p2 = idpool.NewReusableIdPool()

func BenchmarkIncrementalPool(b *testing.B) {
	for n := 0; n < b.N; n++ {
       i := p1.GetId()
       p1.ReleaseId(i)
    }
}

func BenchmarkReusablePool(b *testing.B) {
	for n := 0; n < b.N; n++ {
       i := p2.GetId()
       p2.ReleaseId(i)
    }
}

func TestIncrementalPool(t *testing.T) {

	assert := assert.New(t)
	pool := idpool.NewIncrementalPool()

	assert.Equal(uint64(1), pool.GetId(), "Id should start by 1")
	assert.Equal(uint64(2), pool.GetId(), "IdPool should give incremental ids")
	assert.Equal(uint64(3), pool.GetId(), "IdPool should give incremental ids")
	assert.Equal(uint64(4), pool.GetId(), "IdPool should give incremental ids")

}

func TestReusablePool(t *testing.T) {
	assert := assert.New(t)
	pool := idpool.NewReusableIdPool()

	assert.Equal(uint64(1), pool.GetId(), "Id should start by 1")
	assert.Equal(uint64(2), pool.GetId(), "IdPool should give incremental ids")
	assert.Equal(uint64(3), pool.GetId(), "IdPool should give incremental ids")
	assert.Equal(uint64(4), pool.GetId(), "IdPool should give incremental ids")

	pool.ReleaseId(2)
	assert.Equal(uint64(2), pool.GetId(), "IdPool should reuse id")
	assert.Equal(uint64(5), pool.GetId(), "IdPool should give incremental ids")

	pool.ReleaseId(1)
	pool.ReleaseId(4)
	pool.ReleaseId(2)
	pool.ReleaseId(5)

	assert.Equal(uint64(1), pool.GetId(), "IdPool should reuse ids in a FIFO")
	assert.Equal(uint64(4), pool.GetId(), "IdPool should reuse ids in a FIFO")
	assert.Equal(uint64(2), pool.GetId(), "IdPool should reuse ids in a FIFO")
	assert.Equal(uint64(5), pool.GetId(), "IdPool should reuse ids in a FIFO")
}

