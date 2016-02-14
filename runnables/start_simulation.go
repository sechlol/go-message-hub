package main

import(
	"os"
	"log"
	"flag"
	"sync"
	"time"
	"os/signal"
	"github.com/sech90/go-message-hub/client"
	"github.com/sech90/go-message-hub/message"
	"github.com/sech90/go-message-hub/syncmap"
	"github.com/sech90/go-message-hub/statbucket"
	"github.com/sech90/go-message-hub/testutils"
)

const(
	Address = "localhost"
	Port 	= 9999
	payloadBytes = 1024 * 10
	numClients = 100
	numMessages = 100
	sendInterval = 1
)

var(
	addr string	
	port int
	pSize int	
	cliNum int	
	mexNum int
	interval time.Duration	
	showStat bool
)

var RelayMessage message.Message
var allStats *statbucket.StatBucket
var climap = syncmap.NewSyncMap()
var allCliId []uint64
var payload []byte
var messagesToRead int
var quitApp = make(chan bool)

var startTime time.Time
var endTime time.Time

var outMex, inMex, inByte, outByte int
func main() {

	a 	:= flag.String("addr", Address, "address to connect")
	p 	:= flag.Int("port", Port, "Define port number")
	s 	:= flag.Int("size", payloadBytes, "Size of the payload to send [0-1024000]")
	n 	:= flag.Int("ncli", numClients, "Clients number to run the simulation [0-255]")
	m	:= flag.Int("nmex", numMessages, "Messages to send per client")
	i 	:= flag.Int("i", sendInterval, "time interval between messages in milliseconds")
	ss 	:= flag.Bool("stat", true, "Show statistics report at termination")
	
	flag.Parse()
	addr = *a
	port = *p
	pSize = *s
	cliNum = *n
	mexNum = *m

	interval = time.Duration(*i) * time.Millisecond
	showStat = *ss

	if pSize > 1024*1000 || pSize < 0{
		log.Fatalln("size must be between 0-1024000")
	}

	if cliNum > 255 || cliNum < 0{
		log.Fatalln("clients must be between 0-255")
	}

	if *i < 0{
		log.Fatalln("i must be >= 0")
	}

	log.Println("Initializing...")
	Init()

	log.Println("Connecting",cliNum,"users")
	ConnectClients()

	log.Println("Clients asking for peer list")
	GetList()

	log.Println("Begin simulation")
	Simulate()
	if showStat {
		logProgress()
	}

	log.Println("Simulation ended, disconnecting clients")
	DisconnectClients()

	if showStat {
		PrintStats()
	}

	close(quitApp)
}

func Init(){
	payload = testutils.GenPayload(pSize)
	messagesToRead = (cliNum-1) * mexNum
	allStats = new(statbucket.StatBucket) 

	outMex = cliNum * mexNum
	inMex = (cliNum-1) * mexNum * cliNum

	//HACK: for convenience and semplicity, we can calculate what will be the dimension of a message from here
	inByte 	= inMex  * (5+pSize)
	outByte = outMex * (5+1+(cliNum*8)+pSize)

	//Needed if the application is forced close by the user
	CatchExit()
}

/* Make 2 goroutines for each user: one for writing and one for reading
 * before starting to write, wait that all clients are ready to read
 * or there may be congestions
 */
func Simulate(){
	startTime = time.Now()
	var cliFinish sync.WaitGroup
	var cliReady  sync.WaitGroup

	cliReady.Add(1)
	cliFinish.Add(climap.Length()*2)

	for _, id := range allCliId {
		item,_ := climap.Get(id)
		cli := item.(*client.Client)

		go readLoop(cli, &cliFinish)
		go writeLoop(cli, &cliFinish, &cliReady)
	}

	if showStat {
		go logActivity()
	}
	//clients are ready for reading
	cliReady.Done()

	//wait clients to terminate
	cliFinish.Wait()
	endTime = time.Now()
}

func readLoop(cli *client.Client, finish *sync.WaitGroup){

	for i:=0; i<messagesToRead; i++ {
		<- cli.IncomingRelay()
		if showStat {
			allStats.IncomingMessages.Increase(1)
		}
	}

	finish.Done()
}

func writeLoop(cli *client.Client, finish *sync.WaitGroup, ready *sync.WaitGroup){
	
	list := cli.List()
	if len(list) > 255{
		list = list[:255]
	}
	binary := message.NewRelayRequest(list, payload).ToByteArray()
	ready.Wait()

	for i:=0; i<mexNum; i++ {
		cli.SendBytes(binary)
		if showStat {
			allStats.OutgoingMessages.Increase(1)
		}
		time.Sleep(interval)
	}
	finish.Done()
}

func ConnectClients(){

	var wg sync.WaitGroup

	wg.Add(cliNum)
	for i:=0; i< cliNum; i++ {
		go func(){

			//create a client and connect to server
			cli	:= client.NewClient()
			cli.Connect(addr,port)

			//wait the server reply
			<- cli.IncomingId()

			//add to map
			climap.Set(cli.Id(), cli)
			
			if showStat {
				allStats.ClientsConnected.Increase(1)
			}
			wg.Done()
		}()
	}

	//wait for all clients to be ready
	wg.Wait()
	allCliId = climap.GetKeys()
}

func GetList(){

	var wg sync.WaitGroup

	//create one list request for all clients
	listReq := message.NewRequest(message.List).ToByteArray()

	wg.Add(cliNum)
	for _,id := range allCliId {
		go func(id uint64){

			//send request for list
			item,_ := climap.Get(id)
			cli := item.(*client.Client)
			cli.SendBytes(listReq)

			//wait server reply
			<- cli.IncomingList()

			wg.Done()
		}(id)
	}

	//wait all clients to complete
	wg.Wait()
}

func DisconnectClients(){

	for _,id := range allCliId {
		item,_ := climap.Get(id)
		cli := item.(*client.Client)
		cli.Disconnect()
		
		if showStat {
			allStats.ClientsDisconnected.Increase(1)
		}
	}
}

func CatchExit(){
	
	systemSignals := make(chan os.Signal, 1)

	//intercept external interreupt signals so we can disconnect the clients correctly
	signal.Notify(systemSignals, os.Interrupt)
	go func(){

		//a way to escape if application terminates correctly
	   select{
		case <- systemSignals:
			DisconnectClients()
			if(showStat){
				PrintStats()
			}
		case <- quitApp:
		}
	    
	    os.Exit(2)
	}()
}

//invoed to display the progression
func logActivity(){

	for{
		select{
		case <-quitApp:
			return

		//show an update every second with the current progress percentage
		case <- time.After(1 * time.Second):
			logProgress()
		}
	}
}

func logProgress(){
	readmex := int(allStats.IncomingMessages.Get())
	writtenmex :=  int(allStats.OutgoingMessages.Get())

	var readperc float32 = (float32(readmex)/float32(inMex))*100.0
	var writeperc float32 = (float32(writtenmex)/float32(outMex))*100.0
	log.Printf("Read progress: %.2f (%d/%d) \t Write progress: %.2f (%d/%d)",readperc, readmex, inMex, writeperc, writtenmex, outMex)
}

func PrintStats(){

	durationTime := endTime.Sub(startTime)

	allStats.TimeAlive.Set(uint64(durationTime.Nanoseconds()))

	//add calculated values for byte read and write
	allStats.ByteWritten.Increase(uint64(outByte))
	allStats.ByteRead.Increase(uint64(inByte))

	log.Println("\n\n*** Execution statistics ***\n")
	log.Println(allStats)
}