package client 

import( 
	"log" 
	"net" 
	"errors" 
	"strconv" 
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/mexsocket"
)

type Client struct {
	id 			uint64

	//underlying message socket.
	socket 		*mexsocket.MexSocket

	//channels for controlling goroutine execution
	quitting		chan bool
	doneQuit		chan error

	//channels for delivering the messages
	incomingId 		chan *message.Answer
	incomingList 	chan *message.Answer
	incomingRelay 	chan *message.Answer
	lastClientList	[]uint64
}

func NewClient() *Client {
	return &Client{
    	quitting: 		make(chan bool),
    	doneQuit: 		make(chan error),
    	incomingId: 	make(chan *message.Answer),
    	incomingList: 	make(chan *message.Answer),
    	incomingRelay: 	make(chan *message.Answer),
	}
}

func (c *Client) IncomingId() <-chan *message.Answer {
	return c.incomingId
}

func (c *Client) IncomingList() <-chan *message.Answer {
	return c.incomingList
}

func (c *Client) IncomingRelay() <-chan *message.Answer {
	return c.incomingRelay
}

func (c *Client) QuitChan() <-chan bool {
	return c.quitting
}

func (c *Client) Id() uint64 {
	return c.id
}

func (c *Client) List() []uint64 {
	return c.lastClientList
}

func (c *Client) Connect(address string, port int) error {

	//already connected or pending connection
	if c.socket != nil {
		return errors.New("client connection already open")
	}

	//dial the connection to the server
	conn, err := net.Dial("tcp", address+":"+strconv.Itoa(port)) 
	if err != nil {
		return err
	}

	//enable connection on client
	c.socket = mexsocket.New(0,conn)

	go c.handleConnection()

	idMex := message.NewRequest(message.Identity)
	_, err = c.socket.Send(idMex)
	return err
}

func (c *Client) Send(mex *message.Request) error {

	//already connected or pending connection
	if c.socket == nil {
		return errors.New("client is not connected")
	}

	_, err := c.socket.Send(mex)
	return err
}

func (c *Client) SendBytes(payload []byte) error {

	//already connected or pending connection
	if c.socket == nil {
		return errors.New("client is not connected")
	}

	_, err := c.socket.WriteBytes(payload)
	return err
}

func (c *Client) Disconnect() {

    close(c.quitting)
    c.socket.Close()

    <- c.socket.QuitChan()
    <- c.doneQuit
}

func (c *Client) handleConnection(){

	go c.socket.StartReadService(mexsocket.ModeClient)
	go c.socket.StartWriteService()

	incoming := c.socket.Incoming()
	errChan  := c.socket.ErrorChan()

	for {

		select{
        	//got new message from server
        	case answer,ok := <- incoming:    
        		if ok {
	        		//don't block on forwarding messages to outgoing channel
	        		go c.queueAnswer(answer.(*message.Answer))
	        	}

			/* if socket gives error, print it... at least! There should be a proper error handling, a mechanism
	         * to recover the connection or eventual special errors.. Need more requirements..
	         */
			case err := <- errChan:
				log.Println("received error",err)

			//call disconnect
			case <-c.quitting:  		
	            c.socket.Close()		
	            c.doneQuit <- nil
	            return
      
		}
	}
}

func (c *Client) queueAnswer(ans *message.Answer){

	var ch chan *message.Answer

	switch ans.MexType{
	case message.Identity:
		c.id = ans.Id()
		ch = c.incomingId
	case message.List:
		c.lastClientList = ans.List()
		ch = c.incomingList
	case message.Relay:
		ch = c.incomingRelay
	default:
		log.Println("Client",c.Id(),"Received unknown answer",ans.MexType)
		return
	}

	//Exit if client closes
	select{
		case ch <- ans:
			return
		case <- c.quitting:
			return
	}
}