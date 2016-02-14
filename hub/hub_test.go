package hub_test

import(
	"net"
	"sync"
	"strconv"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sech90/go-message-hub/hub"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/mexsocket"
	"github.com/sech90/go-message-hub/testutils"
)

const(
	addr 	 = "localhost"
	port 	 = 9999
	cliNum 	 = 25
	bodySize = 1024 * 1
	messagesPerCLient = 10
)

/* Fake structure to mock a client */
type Cli struct{
	id 		uint64
	peers	[]uint64
	socket 	*mexsocket.MexSocket
}

var server *hub.Hub
var cliList map[uint64]*Cli
var testBody = make([]byte,bodySize)


func TestInit(t *testing.T){

	cliList = make(map[uint64]*Cli)
	server = hub.NewHub(port)
	go server.Run()

}

func TestCliConnect(t *testing.T){
	assert := assert.New(t)

	ans 	:= new(message.Answer)
	idReq 	:= message.NewRequest(message.Identity)

	for i:= 0; i<cliNum; i++{
		conn, err := net.Dial("tcp", addr+":"+strconv.Itoa(port)) 
		
		assert.Nil(err, "Should be able to connect")
		assert.NotNil(conn, "Connection should be valid")
		
		socket := mexsocket.New(0,conn)		

		socket.Send(idReq)
		socket.Read(ans)
		assert.Equal(message.Identity, ans.MexType, "answer type should be Identity")

		id := ans.Id()
		socket.Id = id

		_, ok := cliList[id]

		assert.False(ok, "Id should not be duplicated")

		cliList[id] = &Cli{id, make([]uint64,0), socket}
		ans.Clear()
		
	}
}

func TestList(t *testing.T){

	assert := assert.New(t)
	var wg sync.WaitGroup

	listReq := message.NewRequest(message.List)
		
	wg.Add(cliNum)
	for _,v := range cliList {
		go func(cli *Cli){

			cli.socket.Send(listReq)

			ans := new(message.Answer)
			_,err := cli.socket.Read(ans)
			assert.Nil(err, "Answer should be valid")
			assert.Equal(message.List, ans.MexType, "Answer type should be List")

			cli.peers = ans.List()

			assert.NotNil(cli.peers, "List should be valid")
			assert.Len(cli.peers, cliNum-1,  "List should contain all the clients ids except for the current")
			assert.False(testutils.IsInList(cli.id, ans.List()), "List should contain all the clients ids except for the current")

			wg.Done()
		}(v)
	}

	wg.Wait()
}

func TestRelay(t *testing.T){

	assert := assert.New(t)
	var wg sync.WaitGroup
	
	wg.Add(cliNum * 2)
	for _,v := range cliList {
		go func(cli *Cli){

			go func(){
				mex := message.NewRelayRequest(cli.peers, testBody).ToByteArray()
				for i:=0; i<messagesPerCLient; i++ {
					_, err := cli.socket.WriteBytes(mex)
					assert.Nil(err, "Write shouldn't fail")
				}
				wg.Done()
			}()

			go func(){
				ans := new(message.Answer)
				var totalMex = (cliNum-1) * messagesPerCLient
				var i = 0
				defer wg.Done()
				for ; i<totalMex; i++ {
					n,err := cli.socket.Read(ans)
					if(n==0 || err != nil){
						break
					}
					assert.Nil(err, "Read shouldn't fail")
					assert.Equal(message.Relay, ans.MexType, "type should be Relay")
					assert.Nil( testutils.CompareBytes(testBody, ans.Body()), "body should be the same")
				}

				assert.Equal(totalMex, i, "Message count mismatch")
			}()
		}(v)
	}
	wg.Wait()
}

func TestDisconnect(t *testing.T){
	
	server.Stop()
	
	for _,v := range cliList {
		ans := new(message.Answer)
		_,err := v.socket.Read(ans)
		assert.NotNil(t,err,"write whould give error when server is closed")
	}
}



