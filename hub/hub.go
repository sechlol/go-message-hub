package hub 

import( 
	"log"
	"net" 
	"strconv" 
	"gopkg.in/fatih/set.v0"

	"github.com/sech90/go-message-hub/hub/idpool"
	"github.com/sech90/go-message-hub/syncmap"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/mexsocket"
)

/* main server and dispatcher for messages */
type Hub struct{
	listener 	net.Listener

	//thread safe map for clients
	socketMap 	*syncmap.SyncMap

	//thread safe id pool for new clients
	idPool 		idpool.IdPool

	//thread safe set for caching clients ID
	idSet		*set.Set

	//used to signal all goroutines to close at once
	quit		chan bool
}

func NewHub(port int) *Hub{

	// listen on all interfaces
	ls, err := net.Listen("tcp", ":"+strconv.Itoa(port))  

	log.Printf("Server begin listen at port: %d", port)
	
	//chec for connection error
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	hub := &Hub{

		listener: 	ls,
		quit: 		make(chan bool),
		idPool: 	idpool.NewReusableIdPool(),
		idSet: 		set.New(),
		socketMap: syncmap.NewSyncMap(),
	}

	return hub

}

func (hub *Hub) Run() error {

	//main server loop
	for{
		//accept new connections
		conn, err := hub.listener.Accept()

		if err != nil {
			hub.Stop()
			return err
		}

		//manage connection
		go hub.handleConnection(conn)
	}
} 

func (hub *Hub) Stop() {

	cliList := hub.idSet.List()
	var id uint64
	var val interface{} 
	var ok bool

	//disconnect all connected clients
	for _, val = range cliList {
		id 	= val.(uint64)
		val, ok = hub.socketMap.Get(id)

		if ok{
			val.(*mexsocket.MexSocket).Close()
		}
	}
	hub.listener.Close() 

}

//Broadcast the message to the clients with id contained in the list
func (hub *Hub) Multicast(ids []uint64, mex *message.Answer){

	//convert the message once
	bytes := mex.ToByteArray()
	for _, id := range ids {
		
		//get client from map
		s, ok := hub.socketMap.Get(id)

		if ok == true {
			//send the data directly to client's write goroutine
			s.(*mexsocket.MexSocket).OutgoingBinary() <- bytes
		}
	}

}

func (hub *Hub) handleConnection(conn net.Conn){

	//get an id from pool
	id := hub.idPool.GetId()

	//create a new socket
	s := mexsocket.New(id, conn)

	//enable socket service goroutines
	go s.StartReadService(mexsocket.ModeServer)
	go s.StartWriteService()

	//add socket to map
	hub.socketMap.Set(id, s)

	//add id to Set
	hub.idSet.Add(id)

	for {
		select{

			//received a request to process
			case mex, ok := <- s.Incoming():
				if ok {
					go hub.processRequest(s, mex.(*message.Request))
				}

			//client is closing. Terminate loop
			case <- s.QuitChan():

				//remove client info from structures
				hub.socketMap.Remove(id)
				hub.idSet.Remove(id)
				return
		}
	}
}

func (hub *Hub) processRequest(socket *mexsocket.MexSocket, req *message.Request){

	if req == nil {
		return
	}

	switch req.MexType {

	//create a new answer and send it over the channel
	case message.Identity:
		answer := message.NewAnswerIdentity(socket.Id)
		socket.Send(answer)

	//get the list of connected clients, remove the current one and send it over the channel
	case message.List:

		//the resulting list is []interface{}
		list := set.Difference(hub.idSet, set.New(socket.Id)).List()

		//create new answer
		answer := new(message.Answer) 
		answer.MexType = message.List
		answer.Payload = convertSetList(list)
		socket.Send(answer)

	case message.Relay:
		
		//create an answer containing the payload
		answer := message.NewAnswerRelay(req.Body)

		//call hub to send the message to the list of clients provided
		hub.Multicast(req.Receivers, answer)
	}
}

func convertSetList(list []interface{}) []byte {

	out := make([]byte,0,len(list)*8)
	
	//temporary container for conversion uint64 --> [8]byte
	buff := make([]byte,8,8)

	//one by one convert the receiver ID into bytearray and append it to the byte message
	for _, el := range list {

		//BigEndian implementation ensures a 8-byte conversion
		message.Uint64ToByteArray(buff, el.(uint64))
		out = append(out, buff...)	
	}

	return out

}