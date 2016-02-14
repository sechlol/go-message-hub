package mexsocket

import(
	"io"
	"net"
	"sync"
	"errors"
	"github.com/sech90/go-message-hub/message"
)

const (
	HEADER_SIZE = 4

	ModeServer = 1
	ModeClient = 2
)

/* A socket for messages and binary data.
 * It can be used syncronously with with the functions Read() Send() ReadByres() WriteBytes()
 * Or it can work asynchronouslty with StartReadService() and StartWriteService()
 * In the latter case, it communicates using the incoming and outgoing channels
 */
type MexSocket struct {
	Id 		 	uint64 
	conn	 	net.Conn
	
	isClosed 	bool
	quitChan 	chan bool
	outgoingBin	chan []byte
	quitDone 	chan error
	errChan 	chan error
	incoming 	chan message.Message
	outgoing 	chan message.Message

	lock 	 	sync.RWMutex
}

func New(id uint64, conn net.Conn) *MexSocket {

	cli := &MexSocket{
		Id: id,
		conn: conn,

		isClosed: false,
		quitChan: 	 make(chan bool),
		outgoingBin: make(chan []byte),
		quitDone: 	 make(chan error),
		errChan: 	 make(chan error),
		incoming: 	 make(chan message.Message),
		outgoing: 	 make(chan message.Message),
	}

	return cli
}

func (s *MexSocket) StartReadService(mode int) {

	var mex message.Message
	defer close(s.incoming)

	for {
		select{
		case <- s.quitChan:
			return
		default:
			if mode == ModeServer {
				mex = message.NewRequestEmpty()
			} else if (mode == ModeClient){
				mex = message.NewAnswerEmpty()
			} else {
				break
			}

			n, lastErr := s.Read(mex)
			
			if lastErr == nil && n > 0{
				s.incoming <- mex
			} else if lastErr == io.EOF {
				go s.Close()
				return
			} else {
				go s.queueErr(lastErr)
			}
		}
	}
}

func (s *MexSocket) StartWriteService() {
	var err error
	for {
		err = nil
		select{
		case <- s.quitChan:
			return
		case mex := <- s.outgoing:
			_,err = s.Send(mex)
		case binary := <- s.outgoingBin:
			_,err = s.WriteBytes(binary)
		}
		if err != nil {
			go s.queueErr(err)
		}
	}
}


func (s *MexSocket) queueErr(err error){
	select{
	case s.errChan <- err:
	case <- s.quitChan:
	}
}

func (s *MexSocket) Incoming() <- chan message.Message {
	return s.incoming
}

func (s *MexSocket) Outgoing() chan <- message.Message {
	return s.outgoing
}

func (s *MexSocket) OutgoingBinary() chan <- []byte {
	return s.outgoingBin
}

func (s *MexSocket) QuitChan() <- chan bool {
	return s.quitChan
}

func (s *MexSocket) ErrorChan() <- chan error {
	return s.errChan
}

func (s *MexSocket) Send(mex message.Message) (int, error) {

	if s.IsClosed() {
		return 0, errors.New("Impossible to write on closed socket")
	}

	//convert message to byte array
	bytes := mex.ToByteArray() 

	//then call the low level writebytes
	n, err := s.WriteBytes(bytes)
	return n, err
}

func (s *MexSocket) Read(mex message.Message) (int, error) {
	
	if s.IsClosed() {
		return 0, errors.New("Impossible to read from closed socket")
	}

	//read bytes 
	bytes, n, err := s.ReadBytes()
	
	if err != nil {
		return n,err
	}

	//build a message from the byte read
	err = mex.FromByteArray(bytes)
	return n, err		
}
	
/* Given a byte array, add a header containing its length */
func (s *MexSocket) WriteBytes(byteMex []byte) (int, error) {

	if s.IsClosed() {
		return 0, errors.New("Impossible to write on closed socket")
	}

	//convert the size into a byte array
	mexHeader := make([]byte,HEADER_SIZE)
	message.Uint32ToByteArray(mexHeader, uint32(len(byteMex)))

	//compose the full arra to write
	toWrite := append(mexHeader, byteMex...)

	var(
		toWriteLen = len(toWrite)
		err error
		totalWritten = 0
		n int
	)

	//iteratively write all the data, until it's finished or there's an error
	for totalWritten < toWriteLen && err == nil {
		n, err = s.conn.Write(toWrite[totalWritten:])
		totalWritten += n
	}
	
	// Return the bytes written, any error
	return totalWritten, err
}

/* Reads the header and then all the rest of the packet */
func (s *MexSocket) ReadBytes() ([]byte, int, error) {
	
	//error if null or empty data
	if s.conn == nil {
		return nil, 0, errors.New("Reader cannot be nil")
	}
	
	var (
		n 	int
		err error
		totalRead int
		totalReadHeader int
		totalReadMessage int
	)

	mexBuffer := make([]byte, HEADER_SIZE)

	//Read the header size
	for totalReadHeader < HEADER_SIZE && err == nil {		
		n, err = s.conn.Read(mexBuffer[totalReadHeader:])
		totalReadHeader += n
	}

	//return if error
	if err != nil { return nil, totalReadHeader, err}
	
	//convert in integer
	mexSize := int(message.ByteArrayToUint32(mexBuffer))
	
	mexBuffer = make([]byte, mexSize)

	//Read the header size
	for totalReadMessage < mexSize && err == nil {		
		n, err = s.conn.Read(mexBuffer[totalReadMessage:])
		totalReadMessage += n
	}

	totalRead = totalReadMessage + totalReadHeader
	
	//return if error
	if err != nil { return nil, totalRead, err}

	return mexBuffer, totalRead, err
}

func (s *MexSocket) IsClosed() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.isClosed
}

func (s *MexSocket) Close() {
	
	s.lock.Lock()
	defer s.lock.Unlock()

	//close only if not closed already
	if !s.isClosed {
		s.isClosed = true
		close(s.quitChan)
		s.conn.Close()
	}
}