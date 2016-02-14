package message_test

import(
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/testutils"
)


var testId   = uint64(12345)
var testNum  = uint32(12345)
var testList = testutils.GenList(255)
var testBody = testutils.GenPayload(1024 * 1000)
var testListByte = message.Uint64ArrayToByteArray(testList)

var tAns 	 = new(message.Answer)
var tAnsList = message.NewAnswerList(testList)
var tAnsBody = message.NewAnswerRelay(testBody)
var tAnsListBytes = tAnsList.ToByteArray()
var tAnsBodyBytes = tAnsBody.ToByteArray()

var tReq 	 = new(message.Request)
var tReqList = message.NewRequest(message.List)
var tReqBody = message.NewRelayRequest(testList, testBody)
var tReqListBytes = tReqList.ToByteArray()
var tReqBodyBytes = tReqBody.ToByteArray()

var mockRW = testutils.NewMockRW(nil, true)

/* Benchmark List Conversions */
func BenchmarkByteArrayToUint64Array(b *testing.B) {
    for n := 0; n < b.N; n++ {
        message.Uint64ArrayToByteArray(testList)
    }
}

func BenchmarkUint64ArrayToByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        message.ByteArrayToUint64Array(testListByte)
    }
}

/* Benchmark Answer */
func Benchmark_AnswerList_FromByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tAns.FromByteArray(tAnsListBytes)
    }
}


func Benchmark_AnswerList_ToByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tAnsList.ToByteArray()
    }
}

func Benchmark_AnswerRelay_FromByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tAns.FromByteArray(tAnsBodyBytes)
    }
}

func Benchmark_AnswerRelay_ToByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tAnsBody.ToByteArray()
    }
}

/* Benchmark request */
func Benchmark_RequestList_FromByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tReq.FromByteArray(tReqListBytes)
    }
}

func Benchmark_RequestList_ToByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tReqList.ToByteArray()
    }
}

func Benchmark_RequestRelay_FromByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tReq.FromByteArray(tReqBodyBytes)
    }
}

func Benchmark_RequestRelay_ToByteArray(b *testing.B) {
    for n := 0; n < b.N; n++ {
        tReqBody.ToByteArray()
    }
}


func TestConversion(t *testing.T){
	buf := make([]byte,8)
	message.Uint64ToByteArray(buf, testId)
	id := message.ByteArrayToUint64(buf)
	assert.Equal(t, testId, id, "uint64 conversion is wrong!");

	buf = make([]byte,4)
	message.Uint32ToByteArray(buf, testNum)
	n := message.ByteArrayToUint32(buf)
	assert.Equal(t, testNum, n, "uint32 conversion is wrong!");

	a1 := message.Uint64ArrayToByteArray(nil)
	a2 := message.Uint64ArrayToByteArray(make([]uint64,0))
	a3 := message.Uint64ArrayToByteArray(testList)

	assert.Nil(t, a1, "Should return nil with nil as input")
	assert.Len(t, a2, 0, "Should give empty byte list")
	assert.Len(t, a3, len(testList) * 8, "One element takes 8 bytes of space")

	l1 := message.ByteArrayToUint64Array(nil)
	l2 := message.ByteArrayToUint64Array(make([]byte,7))
	l3 := message.ByteArrayToUint64Array(make([]byte,0))
	l4 := message.ByteArrayToUint64Array(a3)

	assert.Nil(t, l1, "Should return nil with nil as input")
	assert.Nil(t, l2, "Should return nil with incorrect length")
	assert.Len(t, l3, 0, "Should give empty byte list")
	assert.Nil(t, testutils.CompareList(l4,testList), "Lists should be the same")
}

func TestCreateEmpty(t *testing.T) {

	req := message.NewRequestEmpty().(*message.Request)
	ans := message.NewAnswerEmpty().(*message.Answer)

	if assert.NotNil(t, req, "Empty Request should not be nil") {
		assert.Equal(t, req.Type(), message.Empty, "MessageType should be Empty" )
	}

	assert.Nil(t, req.Receivers, "Receivers should be nil")
	assert.Nil(t, req.Body, 	 "Body should be nil")

	if assert.NotNil(t, ans, "Empty Answer should not be nil") {
		assert.Equal(t, ans.Type(), message.Empty, "MessageType should be Empty" )
	}

	assert.Nil(t, ans.Payload, 	 "Payload should be nil")
}

func TestCreateIdentity(t *testing.T) {
	req := message.NewRequest(message.Identity)
	ans := message.NewAnswerIdentity(testId)

	if assert.NotNil(t, req, "Identity Request should not be nil") {
		assert.Equal(t, req.Type(), message.Identity, "MessageType should be Identity" )
	}

	assert.Nil(t, req.Receivers, "Receivers should be nil")
	assert.Nil(t, req.Body, 	 "Body should be nil")

	if assert.NotNil(t, ans, "Identity Request should not be nil") {
		assert.Equal(t, ans.Type(), message.Identity, "MessageType should be Identity" )
	}

	assert.Equal(t, testId, ans.Id(), "Id not encoded correctly")
	assert.Nil(t, ans.List(), "List should return invalid value")
	assert.Nil(t, ans.Body(), "Body should return invalid value")

}

func TestCreateList(t *testing.T) {
	req := message.NewRequest(message.List)
	ans := message.NewAnswerList(testList)
	ans2 := message.NewAnswerList(make([]uint64,0))

	if assert.NotNil(t, req, "List Request should not be nil") {
		assert.Equal(t, req.Type(), message.List, "MessageType should be List" )
	}

	assert.Nil(t, req.Receivers, "Receivers should be nil")
	assert.Nil(t, req.Body, 	 "Body should be nil")

	if assert.NotNil(t, ans, "List Answer should not be nil") {
		assert.Equal(t, ans.Type(), message.List, "MessageType should be List" )
	}

	assert.Nil(t, testutils.CompareList(testList, ans.List()), "List not encoded correctly")
	assert.Equal(t, uint64(0), ans.Id(), "ID should return invalid value")
	assert.Len(t, ans.List(), len(testList), "List length shoud match")
	assert.Len(t, ans2.List(), 0, "Empty List should stay empty")
	assert.Nil(t, ans.Body(), "Body should return invalid value")
}

func TestCreateRelay(t *testing.T) {

	req := message.NewRelayRequest(testList, testBody)
	ans := message.NewAnswerRelay(testBody)

	if assert.NotNil(t, req, "Relay Request should not be nil") {
		assert.Equal(t, req.Type(), message.Relay, "MessageType should be Relay" )
	}

	if assert.NotNil(t, req.Receivers, "Receivers should not be nil"){
		if err := testutils.CompareList(testList, req.Receivers); err != nil {
			t.Error(err)
		}
	}

	if assert.NotNil(t, req.Body, "Body should not be nil"){
		assert.Nil(t, testutils.CompareBytes(testBody, req.Body), "Body should be the same of the one provided")
	}

	
	assert.Nil(t, testutils.CompareBytes(testBody, ans.Body()), "Body should be the same of the one provided")
	assert.Equal(t, uint64(0), ans.Id(), "ID should return invalid value")
	assert.Nil(t, ans.List(), "List should return invalid value")
	
}

func TestAnswerConversion(t *testing.T){
	
	a1 := message.NewAnswerIdentity(testId)
	a2 := message.NewAnswerList(testList)
	a3 := message.NewAnswerRelay(testBody)
	
	b1 := a1.ToByteArray()
	b2 := a2.ToByteArray()
	b3 := a3.ToByteArray()

	c1 := new(message.Answer)
	c2 := new(message.Answer)
	c3 := new(message.Answer)
	c4 := new(message.Answer)
	
	e1 := c1.FromByteArray(b1)
	e2 := c2.FromByteArray(b2)
	e3 := c3.FromByteArray(b3)
	e4 := c4.FromByteArray(nil)
	
	assert.Nil(t, e1, "No error in conversion (1)")
	assert.Nil(t, e2, "No error in conversion (2)")
	assert.Nil(t, e3, "No error in conversion (3)")
	assert.NotNil(t, e4, "Conversion from nil throws an error (4)")
	
	assert.Nil(t, testutils.CompareAnswer(a1, c1), "Identity answer conversion wrong")
	assert.Nil(t, testutils.CompareAnswer(a2, c2), "List answer conversion wrong")
	assert.Nil(t, testutils.CompareAnswer(a3, c3), "Relay answer conversion wrong")

}



func TestRequestConversion(t *testing.T){

	assert := assert.New(t)

	m1 := message.NewRequest(message.Identity)
	m2 := message.NewRequest(message.List)
	m3 := message.NewRelayRequest(testList, testBody)
	m4 := message.NewRelayRequest(testutils.GenList(message.MAX_RECEIVERS+1), testBody)
	m5 := message.NewRelayRequest(testList, testutils.GenPayload(message.MAX_PAYLOAD+1))

	assert.Nil(m4, "Messages with too many receivers not allowed")
	assert.Nil(m5, "Messages with payload too big not allowed")

	b1  := m1.ToByteArray()
	b2  := m2.ToByteArray()
	b3 	:= m3.ToByteArray()

	c1 := new(message.Request)
	c2 := new(message.Request)
	c3 := new(message.Request)
	c4 := new(message.Request)
	
	e1 := c1.FromByteArray(b1)
	e2 := c2.FromByteArray(b2)
	e3 := c3.FromByteArray(b3)
	e4 := c4.FromByteArray(nil)
	e5 := c4.FromByteArray(make([]byte,0))
	
	assert.Nil(e1, "No error in conversion (1)")
	assert.Nil(e2, "No error in conversion (2)")
	assert.Nil(e3, "No error in conversion (3)")
	assert.NotNil(e4, "Conversion from nil throws an error (4)")
	assert.NotNil(e5, "Conversion from empty throws an error (4)")
	
	assert.Nil(testutils.CompareRequests(m1, c1), "Identity request conversion wrong")
	assert.Nil(testutils.CompareRequests(m2, c2), "List request conversion wrong")
	assert.Nil(testutils.CompareRequests(m3, c3), "Relay request conversion wrong")	
}

func TestClearAnswer(t *testing.T){

	a := message.NewAnswerRelay(testBody)
	a.Clear()

	assert.Equal(t, message.Empty, a.Type(), "Cleared message type should be Empty")
	assert.Nil(t, a.Payload, "Cleared Payload should be Nil")
}

func TestClearRequest(t *testing.T){

	m := message.NewRelayRequest(testList, testBody)
	m.Clear()

	assert.Equal(t, m.Type(), message.Empty, "Cleared message type should be Empty")
	assert.Nil(t, m.Receivers, "Cleared Receivers should be Nil")
	assert.Nil(t, m.Body, "Cleared Body should be Nil")
}