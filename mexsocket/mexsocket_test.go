package mexsocket_test

import (
	"log"
	"net"
	"sync"
	"time"
	"strconv"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/testutils"
	"github.com/sech90/go-message-hub/mexsocket"
)

const(
	addr = "localhost"
	port = 9999

	ListLength = 255
	delayTime = time.Millisecond * 0
	PayloadLength = 1024 * 1000
	SizeTestList = 1
)

var testId 	 = uint64(12345)
var testList = testutils.GenList(ListLength)
var testBody = testutils.GenPayload(PayloadLength)

var tAnsId 	 = message.NewAnswerIdentity(testId)
var tAnsList = message.NewAnswerList(testList)
var tAnsBody = message.NewAnswerRelay(testBody)
var emptyAns = new(message.Answer)

var tReqId 	 = message.NewRequest(message.Identity)
var tReqList = message.NewRequest(message.List)
var tReqBody = message.NewRelayRequest(testList, testBody)
var emptyReq = new(message.Request)

var listener net.Listener
var newConn = make(chan net.Conn)

//for testing
var s1, s2 *mexsocket.MexSocket

//for benchmarks
var bServ, bCli *mexsocket.MexSocket


func BenchmarkSendRequest1(b *testing.B){
	for n := 0; n < b.N; n++ {
		bCli.Outgoing() <- tReqId
		<- bServ.Incoming()
    }
}

func BenchmarkSendRequest2(b *testing.B){
	for n := 0; n < b.N; n++ {
		bCli.Outgoing() <- tReqList
		<- bServ.Incoming()
    }
}

func BenchmarkSendRequest3(b *testing.B){
	for n := 0; n < b.N; n++ {
        bCli.Outgoing() <- tReqBody
		<- bServ.Incoming()
    }
}

func BenchmarkSendAnswer1(b *testing.B){
	for n := 0; n < b.N; n++ {
       bServ.Outgoing() <- tAnsId
		<- bCli.Incoming()
    }
}

func BenchmarkSendAnswer2(b *testing.B){
	for n := 0; n < b.N; n++ {
       bServ.Outgoing() <- tAnsList
		<- bCli.Incoming()
    }
}

func BenchmarkSendAnswer3(b *testing.B){
	for n := 0; n < b.N; n++ {
       bServ.Outgoing() <- tAnsBody
		<- bCli.Incoming()
    }
}

/**/
var bId = tReqId.ToByteArray()
func BenchmarkSendRequestBinary1(b *testing.B){
	for n := 0; n < b.N; n++ {
		bCli.OutgoingBinary() <- bId
		<- bServ.Incoming()
    }
}

var bList = tReqList.ToByteArray()
func BenchmarkSendRequestBinary2(b *testing.B){
	for n := 0; n < b.N; n++ {
		bCli.OutgoingBinary() <- bList
		<- bServ.Incoming()
    }
}

var bBody = tReqBody.ToByteArray()
func BenchmarkSendRequestBinary3(b *testing.B){
	for n := 0; n < b.N; n++ {
        bCli.OutgoingBinary() <- bBody
		<- bServ.Incoming()
    }
}

var aId = tAnsId.ToByteArray()
func BenchmarkSendAnswerBinary1(b *testing.B){
	for n := 0; n < b.N; n++ {
       bServ.OutgoingBinary() <- aId
		<- bCli.Incoming()
    }
}

var aList = tAnsList.ToByteArray()
func BenchmarkSendAnswerBinary2(b *testing.B){
	for n := 0; n < b.N; n++ {
       bServ.OutgoingBinary() <- aList
		<- bCli.Incoming()
    }
}

var aBody = tAnsBody.ToByteArray()
func BenchmarkSendAnswerBinary3(b *testing.B){
	for n := 0; n < b.N; n++ {
       bServ.OutgoingBinary() <- aBody
		<- bCli.Incoming()
    }
}

func TestInit(t *testing.T){
	
	listener, _ := net.Listen("tcp", ":"+strconv.Itoa(port)) 

	go func(){

		defer close(newConn)
		for{
			conn, err := listener.Accept()
			if err != nil {
				log.Println("return",err)
				return
			}
			newConn <- conn
		}
	}()

	s1, s2 = giveCoupledSockets(1,2)
	bServ, bCli = initBenchmark(1111,2222)

	assert.NotNil(t, s1, "Connection should be enstablished")
	assert.NotNil(t, s2, "Connection should be enstablished")
	
}

func TestEndpointsRequest(t *testing.T){ 
	req := new(message.Request)

	var tests = loadList(tReqBody, SizeTestList)

    for _, test := range tests {
		_, e1 := s1.Send(test)
		assert.Nil(t, e1, "Send request should work")

		_, e2 := s2.Read(req)
		assert.Nil(t, e2, "Read request should work")
		assert.Nil(t, testutils.CompareRequests(test.(*message.Request), req), "Requests should be the same")
	}
}

func TestEndpointsRequestGoroutine(t *testing.T){ 
	
	var wg sync.WaitGroup
	var tests = loadList(tReqBody, SizeTestList)

	wg.Add(2)
	go func(){
		for _, test := range tests {
			_, e1 := s1.Send(test)
			assert.Nil(t, e1, "Send request should work")
		}
		wg.Done()
	}()

	go func(){
		for _,_ = range tests{
			req := new(message.Request)
			_, e2 := s2.Read(req)
			assert.Nil(t, e2, "Read request should work")

			if req.MexType == message.Identity {
				assert.Nil(t, testutils.CompareRequests(tReqId, req), "Requests should be the same")
			} else if req.MexType == message.List {
				assert.Nil(t, testutils.CompareRequests(tReqList, req), "Requests should be the same")
			} else if req.MexType == message.Relay {
				assert.Nil(t, testutils.CompareRequests(tReqBody, req), "Requests should be the same")
			} else {
				t.Error("mextype unknown")
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestEndpointsServices(t *testing.T){ 

	go s1.StartReadService(mexsocket.ModeServer)
	go s1.StartWriteService()

	go s2.StartReadService(mexsocket.ModeClient)
	go s2.StartWriteService()


	go func(){s2.Outgoing() <- tReqBody}()
	req := <- s1.Incoming()
	assert.Nil(t, testutils.CompareRequests(tReqBody, req.(*message.Request)), "request should be received correctly")

	go func(){s1.Outgoing() <- tAnsBody}()
	ans := <- s2.Incoming()
	assert.Nil(t, testutils.CompareAnswer(tAnsBody, ans.(*message.Answer)), "answer should be received correctly")
}

func TestDisconnect(t *testing.T){
	assert := assert.New(t)

	assert.False(s1.IsClosed(), "Socket should be open")
	assert.False(s2.IsClosed(), "Socket should be open")

	go s1.Close()
	<- s1.QuitChan()
	assert.True(s1.IsClosed(), "Socket should be closed")

	mex, ok := <- s1.Incoming()
	assert.Nil(mex, "Socket should be closed")
	assert.False(ok, "Socket should be closed")

	_, e1 := s1.Send(tAnsId)
	_, e2 := s1.Read(tAnsId)
	assert.NotNil(e1, "write on closed socket not possible")
	assert.NotNil(e2, "read from closed socket not possible")

	<- s2.QuitChan()
	_, e1 = s2.Send(tAnsId)
	_, e2 = s2.Read(tAnsId)
	assert.True(s2.IsClosed(), "closing an endpoint should have effect on the other")
	assert.NotNil(e1, "closing an endpoint should have effect on the other")
	assert.NotNil(e2, "closing an endpoint should have effect on the other")

}

func loadList(m message.Message, size int) []message.Message {
	out := make([]message.Message,size)
	for i:=0; i<size; i++ {
		out[i] = m
	}
	return out
}

func initBenchmark(id1, id2 uint64) (*mexsocket.MexSocket, *mexsocket.MexSocket) {
	s, c := giveCoupledSockets(id1,id2)

	go s.StartReadService(mexsocket.ModeServer)
	go s.StartWriteService()

	go c.StartReadService(mexsocket.ModeClient)
	go c.StartWriteService()

	return s, c
}

func giveCoupledSockets(id1, id2 uint64) (*mexsocket.MexSocket, *mexsocket.MexSocket){

	conn1, err := net.Dial("tcp", addr+":"+strconv.Itoa(port)) 
	conn2 := <- newConn

	if err != nil || conn2 == nil {
		return nil, nil
	}

	s1 := mexsocket.New(id1, conn1)
	s2 := mexsocket.New(id2, conn2)

	return s1, s2
}