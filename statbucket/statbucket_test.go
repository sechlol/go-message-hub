package statbucket_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/statbucket"
)


func TestStat(t *testing.T){

	var (
		a uint64 = 10
		b uint64 = 5
	)

	assert := assert.New(t)
	var stat statbucket.Stat

	stat.Set(a)
	assert.Equal(a, stat.Get(), "Values should be the same")

	stat.Increase(b)
	assert.Equal(a+b, stat.Get(), "Should increase the value")

}

var b1, b2 statbucket.StatBucket
func TestBucket(t *testing.T){
	assert := assert.New(t)

	b1.TimeAlive.Set(1)
	b1.ClientsConnected.Set(2)
	b1.ClientsDisconnected.Set(3)
	b1.IncomingMessages.Set(4)
	b1.OutgoingMessages.Set(5)
	b1.ByteRead.Set(6)
	b1.ByteWritten.Set(7)

	b2.TimeAlive.Set(7)
	b2.ClientsConnected.Set(6)
	b2.ClientsDisconnected.Set(5)
	b2.IncomingMessages.Set(4)
	b2.OutgoingMessages.Set(3)
	b2.ByteRead.Set(2)
	b2.ByteWritten.Set(1)

	b1.Merge(&b2)
	assert.Equal(uint64(1), b1.TimeAlive.Get(), "TimeAlive is not affacted by merge")
	assert.Equal(uint64(8), b1.ClientsConnected.Get(), "Merge")
	assert.Equal(uint64(8), b1.ClientsDisconnected.Get(), "Merge")
	assert.Equal(uint64(8), b1.IncomingMessages.Get(), "Merge")
	assert.Equal(uint64(8), b1.OutgoingMessages.Get(), "Merge")
	assert.Equal(uint64(8), b1.ByteRead.Get(), "Merge")
	assert.Equal(uint64(8), b1.ByteWritten.Get(), "Merge")
}

func TestBucketSerialize(t *testing.T){
	assert := assert.New(t)

	arr := b1.ToByteArray()
	b2.FromByteArray(arr)

	assert.Equal(b1.TimeAlive.Get(), b2.TimeAlive.Get(), "Element should be the same")
	assert.Equal(b1.ClientsConnected.Get(), b2.ClientsConnected.Get(), "Element should be the same")
	assert.Equal(b1.ClientsDisconnected.Get(), b2.ClientsDisconnected.Get(), "Element should be the same")
	assert.Equal(b1.IncomingMessages.Get(), b2.IncomingMessages.Get(), "Element should be the same")
	assert.Equal(b1.OutgoingMessages.Get(), b2.OutgoingMessages.Get(), "Element should be the same")
	assert.Equal(b1.ByteRead.Get(), b2.ByteRead.Get(), "Element should be the same")
	assert.Equal(b1.ByteWritten.Get(), b2.ByteWritten.Get(), "Element should be the same")

}

func TestString(t *testing.T){
	assert.Equal(t, b1.String(), b2.String(), "Strings should equal")
}