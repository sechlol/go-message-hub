package statbucket

import (
	"fmt"
	"sync"
	"bytes"
	"github.com/sech90/go-message-hub/message"
)

/* very basic thread safe read/write value */
type Stat struct{
	lock sync.RWMutex
	value uint64
}

func (s *Stat) Increase(amount uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.value += amount
}

func (s *Stat) Set(amount uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.value = amount
}

func (s *Stat) Get() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.value
}

/* convenience object that gather some common parameters */
type StatBucket struct{
	TimeAlive			Stat
	ClientsConnected 	Stat
	ClientsDisconnected Stat
	IncomingMessages	Stat
	OutgoingMessages	Stat
	ByteRead			Stat
	ByteWritten			Stat
}

//given another StatBucket, increase all the fields by the other's amount 
func (bucket *StatBucket) Merge(b *StatBucket) {
	bucket.ClientsConnected.Increase(b.ClientsConnected.Get()) 
	bucket.ClientsDisconnected.Increase(b.ClientsDisconnected.Get()) 
	bucket.IncomingMessages.Increase(b.IncomingMessages.Get()) 
	bucket.OutgoingMessages.Increase(b.OutgoingMessages.Get()) 
	bucket.ByteRead.Increase(b.ByteRead.Get()) 
	bucket.ByteWritten.Increase(b.ByteWritten.Get()) 
}

func (bucket *StatBucket) ToByteArray() []byte {
	
	arr := []uint64{
		bucket.TimeAlive.Get(), 
		bucket.ClientsConnected.Get(), 
		bucket.ClientsDisconnected.Get(), 
		bucket.IncomingMessages.Get(), 
		bucket.OutgoingMessages.Get(), 
		bucket.ByteRead.Get(), 
		bucket.ByteWritten.Get(), 
	}

	return message.Uint64ArrayToByteArray(arr)
}


func (bucket *StatBucket) FromByteArray(arr []byte) {
	
	values := message.ByteArrayToUint64Array(arr)

	bucket.TimeAlive.Set(values[0])
	bucket.ClientsConnected.Set(values[1])
	bucket.ClientsDisconnected.Set(values[2])
	bucket.IncomingMessages.Set(values[3])
	bucket.OutgoingMessages.Set(values[4])
	bucket.ByteRead.Set(values[5])
	bucket.ByteWritten.Set(values[6])
}

func (bucket *StatBucket) String() string {
	
	var buffer bytes.Buffer

	cliConn 	:= bucket.ClientsConnected.Get()
	cliDisconn 	:= bucket.ClientsDisconnected.Get()
	incoming 	:= bucket.IncomingMessages.Get()
	outgoing 	:= bucket.OutgoingMessages.Get()
	bytesRead 	:= bucket.ByteRead.Get()
	bytesWrote 	:= bucket.ByteWritten.Get()		
	timeInSeconds := float64(bucket.TimeAlive.Get())/1000000000


    buffer.WriteString("\n")
    buffer.WriteString(fmt.Sprintf("Time alive: %f \n",timeInSeconds))
    buffer.WriteString(fmt.Sprintf("Clients Connected: %d \n",cliConn))
    buffer.WriteString(fmt.Sprintf("Clients Disconnected: %d \n",cliDisconn))
    buffer.WriteString(fmt.Sprintf("Incoming Messages: %d \n",incoming))
    buffer.WriteString(fmt.Sprintf("Outgoing Messages: %d \n",outgoing))
    buffer.WriteString(fmt.Sprintf("Bytes Read: %d \n",bytesRead))
    buffer.WriteString(fmt.Sprintf("Bytes Written: %d \n",bytesWrote))
    
    //avoid division by 0
    if timeInSeconds > 0 {
	    buffer.WriteString("\n*** Average values over lifetime ***\n")
	    buffer.WriteString(fmt.Sprintf("Incoming Messages/Second: %.2f \n",(float64(incoming)/timeInSeconds)))
	    buffer.WriteString(fmt.Sprintf("Outgoing Messages/Second: %.2f \n",(float64(outgoing)/timeInSeconds)))
	    buffer.WriteString(fmt.Sprintf("Bytes Read/Second: %.2f \n",(float64(bytesRead)/timeInSeconds)))
	    buffer.WriteString(fmt.Sprintf("Bytes Written/Second: %.2f \n",(float64(bytesWrote)/timeInSeconds)))
	}
	return buffer.String()
}