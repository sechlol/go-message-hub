package client_test

import( 
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/client"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/testutils"
)

const(
	addr 	= "localhost"
	badAddr = "trollol"
	port 	= 9999
	badPort = 6666
)

var testList = testutils.GenList(1)
var testBody = testutils.GenPayload(1024 * 1)

var server = testutils.NewServer()

func TestInit(t *testing.T){
	server.Run(port, true)
}

func TestConnection(t *testing.T){
	
	var err error
	assert := assert.New(t)

	c := client.NewClient()

	assert.NotNil(c, "Client should be valid")
	assert.Equal(uint64(0), c.Id(), "Id should be not initializd (0)")

	err = c.Connect(badAddr,badPort)
	assert.NotNil(err, "Connection to bad addresses should fail")

	err = c.Send(message.NewRequest(message.Identity))
	assert.NotNil(err, "Send should fail if client is not conneccted")

	err = c.Connect(addr,port)
	assert.Nil(err, "Connection should work")

	id := <- c.IncomingId()
	assert.Equal(id.Id(), c.Id(), "client should receive and set ID")

	err = c.Connect(addr,port)
	assert.NotNil(err, "Connection should refuse if client is already connected")

	c.Disconnect()
}

func TestList(t *testing.T) {
	var err error
	assert := assert.New(t)

	c := client.NewClient()
	c.Connect(addr,port)
	<- c.IncomingId()

	mList := message.NewRequest(message.List)
	err = c.Send(mList)
	assert.Nil(err, "Client should send without errors")

	server.WriteTo(c.Id(), message.NewAnswerList(make([]uint64,0)))
	<- c.IncomingList()
	assert.Len(c.List(), 0, "List length should be 0")
	
	err = c.Send(mList)
	server.WriteTo(c.Id(), message.NewAnswerList(testList))
	<- c.IncomingList()
	assert.Nil(testutils.CompareList(testList, c.List()), "List should match the one from server")
}

func TestRelay(t *testing.T){
	assert := assert.New(t)

	c := client.NewClient()
	c.Connect(addr,port)
	<- c.IncomingId()

	mex := message.NewRelayRequest(testList, testBody)
	err := c.Send(mex)
	assert.Nil(err, "Client should send without errors")

	server.WriteTo(c.Id(), message.NewAnswerRelay(testBody))
	ans := <- c.IncomingRelay()
	assert.Nil(testutils.CompareBytes(testBody, ans.Body()), "Bodies should be the same")

}

func TestTerminate(t *testing.T){
	server.Stop()
}
