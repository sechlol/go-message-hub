package testutils

import(
	"io"
	"os"
	"net"
	"log"
	"fmt"
	"bytes"
	"strconv"
	"crypto/rand"
	"gopkg.in/fatih/set.v0"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/mexsocket"

)
/* A mock reader-writer. It has a memory buffer with the possibililty to cycle it indefinitely
 * to simulate endless inputs
 */
type MockRW struct {
	buf []byte
	cycle bool
}

func NewMockRW(toRead []byte, cycle bool) *MockRW {
	return &MockRW{toRead, cycle}
}

/* Structure to mock a tcp server. It has functions to control the response */
type MockServer struct{
	mockRW		io.ReadWriter
	listener 	net.Listener
	nextId 		uint64
	quit 		chan bool
	quitDone 	chan bool
	Clichan 	chan uint64

	sockets 	map[uint64]*mexsocket.MexSocket
}

func NewServer() *MockServer{
	return &MockServer{
		nextId: 	1,
		Clichan: 	make(chan uint64,1000),
		quit: 		make(chan bool),
		quitDone: 	make(chan bool),
		sockets: 	make(map[uint64]*mexsocket.MexSocket),
	}
}

/* Generate a random list withh the specified length */
func GenList(length int) []uint64 {

	list := make([]uint64,length)
	for i,_ := range list {
		list[i] = uint64(i)
	}
	return list
}

/* Generate a random byte array with the specified length */
func GenPayload(length int) []byte {

	out := make([]byte, length)
	rand.Read(out)
	return out
}

func IsInList(id uint64, list []uint64) bool {

	for _,v := range list {
		if v == id {
			return true 
		}
	}

	return false
}

//use sets to diff lists
func DiffLists(l1, l2 []uint64) []uint64 {

	s1 := set.NewNonTS()
	s2 := set.NewNonTS()

	var v uint64
	for _,v = range l1 {
		s1.Add(v)
	}

	for _,v = range l2 {
		s2.Add(v)
	}

	diff := set.Difference(s1, s2)
	diffList := make([]uint64, diff.Size())

	for i := range diffList {
		diffList[i] = diff.Pop().(uint64)
	}

	return diffList
}

//compare two requests and returns an error if some differences are found
func CompareRequests(m1, m2 *message.Request) error {

	if m1.MexType != m2.MexType {
		return fmt.Errorf("Type mismatch. Expect %d, got %d", m1.MexType, m2.MexType)
	}

	if err := CompareList(m1.Receivers, m2.Receivers); err != nil {
		return err
	}

	if !bytes.Equal(m1.Body, m2.Body) {
		return fmt.Errorf("Body mismatch")
	}

	return nil
}

func CompareAnswer(a1, a2 *message.Answer) error {

	if a1.MexType != a2.MexType {
		return fmt.Errorf("Type mismatch. Expect %d, got %d", a1.MexType, a2.MexType)
	}

	if !bytes.Equal(a1.Payload, a2.Payload) {
		return fmt.Errorf("Payload mismatch")
	}

	return nil
}

func CompareBytes(a, b []byte) error {
	if a == nil && b == nil{
		return nil
	}

	if a == nil || b == nil{
		return fmt.Errorf("List nil test: got %s",b)
	}

	if len(a) != len(b){
		return fmt.Errorf("List length mismatch: expected %d, got %d",len(a), len(b))
	}

	if !bytes.Equal(a,b) {
		return fmt.Errorf("List content mismatch")
	}

	return nil
}

func CompareList(a, b []uint64) error {
	
	if a == nil && b == nil{
		return nil
	}

	if a == nil || b == nil{
		return fmt.Errorf("List nil test: got %s",b)
	}

	if len(a) != len(b){
		return fmt.Errorf("List length mismatch: expected %d, got %d",len(a), len(b))
	}

	for k, v := range a {
		if v != b[k]{
			return fmt.Errorf("List value mismatch: expected[%d] = %d, got %d",k, v, b[k])
		}
	}

	return nil
}

//Read from mock reader-writer
func (r *MockRW) Read(p []byte) (n int, err error) {
	if len(r.buf) == 0 {
		return 0, io.EOF
	}

	limit := len(p)
	if len(r.buf) < limit {
		limit = len(r.buf)
	}

	for i:= 0; i<limit; i++ {
		p[i] = r.buf[i]
	}

	//if cycling is enabled, append the read data at the end
	if(r.cycle){
		r.buf = append(r.buf[limit:], r.buf[:limit]...)
	} else {
		r.buf = r.buf[limit:]
	}
	return limit, nil
}

//write appends to buffer
func (r *MockRW) Write(p []byte) (n int, err error) {
	
	r.buf = append(r.buf, p...)
	return len(p), nil
}

func (r *MockRW) Clear() {
	r.buf = make([]byte,0)
}

func (r *MockRW) Cycle(enable bool) {
	r.cycle = enable
}

//start the mock server. It gets a tcp connection
func (server *MockServer) Run(port int, autoReadRoutine bool) {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))   
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Println("Mock server running on port",port)

	server.listener	= ln

	go func(){
		for {
			select{
			case <- server.quit:
				server.quitDone <- true
				return
			default:
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println(err)
					continue
				}
				s := mexsocket.New(server.nextId, conn)
				server.sockets[server.nextId] = s
				
				//notify that a new client connected
				server.Clichan <- server.nextId

				if(autoReadRoutine){
					go server.newConnection(server.nextId, s)
				}

				server.nextId++
			}
		}
	}()
}

func (server *MockServer) NextId() uint64 {
	return server.nextId
}

func (server *MockServer) Stop() {
	close(server.quit)
	server.listener.Close()
	<- server.quitDone
}

func (server *MockServer) WriteTo(id uint64, mex message.Message) (int, error){
	s,ok := server.sockets[id]
	if ok{
		return s.Send(mex)
	} else {
		log.Fatalln("MocServer id ",id," not present (length:",len(server.sockets))
		return 0, nil
	}
}


func (server *MockServer) ReadAnswerFrom(id uint64) *message.Answer {
	s := server.sockets[id]
	ans := new(message.Answer)
	s.Read(ans)
	return ans
}

func (server *MockServer) ReadRequestFrom(id uint64) *message.Request {
	s := server.sockets[id]
	req := new(message.Request)
	s.Read(req)
	return req
}

func (server *MockServer) newConnection(id uint64, s *mexsocket.MexSocket){

	var req *message.Request
	for {
		req = new(message.Request)
		_,err := s.Read(req)
		
		if err != nil {
			log.Println("MocServer read error: ",err)
			break
		}
		
		if(req.MexType == message.Identity){
			server.WriteTo(id, message.NewAnswerIdentity(id))
		}
	}
}