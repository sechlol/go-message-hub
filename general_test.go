package main_test

import(
	"log"
	"time"
	"sync"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/hub"
	"github.com/sech90/go-message-hub/client"
	"github.com/sech90/go-message-hub/syncmap"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/testutils"
)

const(
	Port = 9999
	Addr = "localhost"

	TimeoutTime 	= 5 * time.Second
	ClientsNum 		= 25
	PayloadBytes 	= 1024 * 10
	MessagesPerCLient = 10
)

var climap 	 = syncmap.NewSyncMap()
var server 	 = hub.NewHub(Port)
var testBody = testutils.GenPayload(PayloadBytes)
var allCliId []uint64

func TestInit(t *testing.T){
	go server.Run()
}

func TestIdentity(t *testing.T){

	assert := assert.New(t)
	var wg sync.WaitGroup

	wg.Add(ClientsNum)
	for i:=0; i< ClientsNum; i++ {
		go func(){
			c 	:= client.NewClient()
			err := c.Connect(Addr,Port)

			assert.Nil(err, "Connection should be valid")

			ans := <- c.IncomingId()
			assert.NotNil(ans, "Should receive an answer from ID channel")
			climap.Set(c.Id(), c)
	
			wg.Done()
		}()
	}
	wg.Wait()
	allCliId = climap.GetKeys()
}

func TestList(t *testing.T){

	assert := assert.New(t)
	var wg sync.WaitGroup

	mex := message.NewRequest(message.List)

	wg.Add(ClientsNum)
	for i:=0; i< ClientsNum; i++ {
		val,_ := climap.Get(allCliId[i])
		go func(cli *client.Client){

			cli.Send(mex)
			ans := <- cli.IncomingList()
			assert.NotNil(ans, "Should receive an answer from List channel")

			assert.False(testutils.IsInList(cli.Id(),cli.List()))

			diff := testutils.DiffLists(allCliId, cli.List())
			assert.Len(diff, 1, "diff list should contain only the client ID")
			assert.Equal(cli.Id(), diff[0], "diff list should contain only the client ID")

			wg.Done()
		}(val.(*client.Client))
	}

	wg.Wait()
}

func TestRelayTwo(t *testing.T){

	if len(allCliId) < 2 {
		return
	}
	
	val,_ := climap.Get(allCliId[0])
	c1 := val.(*client.Client)
	val,_ = climap.Get(allCliId[1])
	c2 := val.(*client.Client)

	var wg sync.WaitGroup
	wg.Add(MessagesPerCLient)
	var counter int

	req := message.NewRelayRequest([]uint64{c2.Id()}, testBody)
	buf := make([]byte,8)
	for i:=0; i<MessagesPerCLient; i++ {

		message.Uint64ToByteArray(buf, uint64(i))
		req.Body = buf
		c1.Send(req)
	}
	
	for i:=0; i<MessagesPerCLient; i++ {
		
		counter++
		wg.Done()
	}

	wg.Wait()
	assert.Equal(t,MessagesPerCLient, counter,"Client2 should have received all messages")

}

func TestRelay(t *testing.T){

	assert := assert.New(t)
	var wg sync.WaitGroup
	var expectedMex int = MessagesPerCLient * (ClientsNum-1)

	wg.Add(ClientsNum)
	for i:=0; i< ClientsNum; i++ {
		
		cli,_ := climap.Get(allCliId[i])

		go func(cli *client.Client){
			received := ForwardAndListen(cli)
			assert.Equal(expectedMex, received, "Client didnt receive the correct number of messages")
			wg.Done()

		}(cli.(*client.Client))
	}

	wg.Wait()
}

func TestDisconnect(t *testing.T){

	for _, id := range allCliId {
		val,_ := climap.Get(id)
		cli := val.(*client.Client)

		cli.Disconnect()
	}
}

func TestPostDisconnect(t *testing.T){
	assert := assert.New(t)

	cli := client.NewClient()
	err := cli.Connect(Addr, Port)
	assert.Nil(err, "new client should connect")

	//wait for id
	<- cli.IncomingId()

	cli.Send(message.NewRequest(message.List))

	//wait for list
	<- cli.IncomingList()

	if len(cli.List()) > 0 {
		time.Sleep(time.Millisecond * 200)
		cli.Send(message.NewRequest(message.List))
		//wait for list
		<- cli.IncomingList()		
	}

	assert.Len(cli.List(), 0, "Server should return empty list")

	cli.Disconnect()
}

func TestEnd(t  *testing.T){
	server.Stop()
}

func ForwardAndListen(cli *client.Client) int {

	var received int
	req := message.NewRelayRequest(cli.List(), testBody)
	incoming := cli.IncomingRelay()
	var expectedIncoming int = MessagesPerCLient * (ClientsNum-1)

	for i:=0; i<MessagesPerCLient; i++ {
		go cli.Send(req)
	}

	for i:=0; i<expectedIncoming; i++ {
		select{
			case <- incoming:
				received++
			case <- time.After( TimeoutTime ):
				log.Println(cli.Id(),"timed out")
				return received
		}
	}

	return received

}