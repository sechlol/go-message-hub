package message

import(
	"errors"
	"encoding/binary"
)

const(
	MAX_PAYLOAD   = int(1024 * 1000)
	MAX_RECEIVERS = int(255)
	CHUNK_LEN 	= 1024 * 2
	HEADER_SIZE = 4

	Empty 		= byte(0)
	Identity 	= byte(1)
	List 		= byte(2)
	Relay 		= byte(3)
	Stat 		= byte(4)
)

type Message interface {
	Clear()
	Type() byte
	ToByteArray() []byte
	FromByteArray(arr []byte) error
}

/* Requests are always CLIENT --> SERVER */
type Request struct {
	MexType byte

	Receivers []uint64
	Body []byte
}

/* Answers are always SERVER --> CLIENT*/
type Answer struct{
	MexType byte
	Payload []byte

	cachedList []uint64
}

/****
* Answer related methods
****/
func NewAnswerEmpty() Message {
	return new(Answer)
}

func NewRequestEmpty() Message {
	return new(Request)
}

func NewAnswerIdentity(id uint64) *Answer {

	payload := make([]byte, 8, 8)

	//convert id in payload 
	Uint64ToByteArray(payload,id)
	return &Answer{Identity, payload, nil}
}

func NewAnswerList(ids []uint64) *Answer {

	//convert ids in payload 
	payload := Uint64ArrayToByteArray(ids)
	return &Answer{List, payload, nil}
}

func NewAnswerRelay(p []byte) *Answer {
	return &Answer{Relay, p, nil}
}

func (a *Answer) Type() byte {
	return a.MexType
}

func (a *Answer) Id() uint64 {

	if a.MexType == Identity && a.Payload != nil && len(a.Payload) > 0 {
		return ByteArrayToUint64(a.Payload)
	}
	return 0
}

func (a *Answer) List() []uint64 {

	if a.cachedList != nil {
		return a.cachedList
	}

	if a.MexType == List && a.Payload != nil && len(a.Payload) > 0 {
		return ByteArrayToUint64Array(a.Payload)
	}
	return nil
}

func (a *Answer) ListAndCache() []uint64 {
	a.cachedList = a.List()
	return a.cachedList
}

func (a *Answer) Body() []byte {
	if a.MexType == Relay && a.Payload != nil && len(a.Payload) > 0 {
		return a.Payload
	}
	return nil
}

func (a *Answer) Clear() {
	a.MexType 	= Empty
	a.Payload 	= nil
	a.cachedList = nil
}

func (a *Answer) ToByteArray() []byte {
	return append([]byte{a.MexType}, a.Payload...)
}

func (a *Answer) FromByteArray(arr []byte) error {
	
	//error if null or empty data
	if arr == nil || len(arr) == 0 {
		return errors.New("Buffer cannot be empty")
	}

	a.MexType = arr[0]
	a.Payload = arr[1:]

	return nil
}

/****
* Request related methods
****/
func NewRequest(reqType byte) *Request {
	return &Request{reqType, nil, nil}
}

func NewRelayRequest(rec []uint64, body []byte) *Request {

	if(len(rec) > 255){
		return nil
	}

	if(len(body) > MAX_PAYLOAD){
		return nil
	}

	return &Request{Relay, rec, body}
}

func (r *Request) Type() byte {
	return r.MexType
}

func (r *Request) Clear() {
	r.MexType 	= Empty
	r.Receivers = nil
	r.Body 		= nil
}

func (r *Request) ToByteArray() []byte {
	
	//for simple messages, only one byte is necessary
	if r.MexType != Relay {
		return []byte{r.MexType}
	}
	
	//calculate total length of output bytearray
	receiversLength := len(r.Receivers)
	dimension 		:= 2 + ( receiversLength * 8 ) + len(r.Body)

	//create array that fits the data exactly 
	arr := make([]byte, 0, dimension)

	//append message type and receivers length (max 255)
	arr = append(arr, r.MexType, byte(receiversLength))
	
	//convert and append list of receivers
	arr = append(arr, Uint64ArrayToByteArray(r.Receivers)...)

	//append the body and return the result
	return append(arr, r.Body...)
}

func (r *Request) FromByteArray(arr []byte) error {

	//error if null or empty data
	if arr == nil || len(arr) == 0 {
		return errors.New("Buffer cannot be empty")
	}

	//assign message type
	r.MexType = arr[0]

	//if the message is not Relay, we're done 
	if(r.MexType != Relay){
		return nil
	}

	//get number of receivers
	receiversLength := int(arr[1])
	
	//index from where the body starts
	startBody := (receiversLength*8)+2

	//create slice for containing receivers
	r.Receivers = ByteArrayToUint64Array(arr[2:startBody])

	//slice the body
	r.Body = arr[startBody:]

	return nil	
}

func Uint64ToByteArray(out []byte, n uint64) {
	binary.BigEndian.PutUint64(out, n)
}

func ByteArrayToUint64( b []byte ) uint64 {
	return binary.BigEndian.Uint64(b)
}

func Uint32ToByteArray(out []byte, n uint32) {
	binary.BigEndian.PutUint32(out, n)
}

func ByteArrayToUint32( b []byte ) uint32 {
	return binary.BigEndian.Uint32(b)
}

func Uint64ArrayToByteArray(intArr []uint64) []byte {

	//return empty slice
	if intArr==nil {
		return nil
	} 

	if len(intArr) == 0 {
		return make([]byte,0)
	}

	out := make([]byte,0,len(intArr)*8)
	
	//temporary container for conversion uint64 --> [8]byte
	byteId := make([]byte,8,8)

	//one by one convert the receiver ID into bytearray and append it to the byte message
	for _, receiverId := range intArr {

		//BigEndian implementation ensures a 8-byte conversion
		Uint64ToByteArray(byteId, receiverId)
		out = append(out, byteId...)	
	}

	return out
}

func ByteArrayToUint64Array(bArr []byte) []uint64 {

	//bArr should be not null and length divisible by 8 
	if bArr == nil || len(bArr) % 8 != 0 {
		return nil
	}

	if len(bArr) == 0 {
		return make([]uint64,0)
	}


	var length int = len(bArr)/8
	var startIndex int
	var endIndex int

	out := make([]uint64, length, length)

	//for all the receivers
	for i:=0; i<length; i++ {

		//calculate indexex in the bytearray
		startIndex 	= (i*8)
		endIndex 	= startIndex + 8

		//convert the 8-byte slice into uint64
		id := ByteArrayToUint64(bArr[startIndex:endIndex])

		//assign to out array
		out[i] = id
	}

	return out
}