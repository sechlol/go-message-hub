package main

import(
	"os"
	"log"
	"flag"
	"os/signal"
	"github.com/sech90/go-message-hub/hub"
	"github.com/sech90/go-message-hub/statbucket"
)

const(
	Port = 9999
)

func main() {

	port := flag.Int("port", Port, "Define port number")
	showStats := flag.Bool("stat", true, "Show statistics report at termination")

	flag.Parse()

	hub := hub.NewHub(*port)
	ServerConnect(hub, *showStats)

}

func ServerConnect(hub *hub.Hub, showStats bool){
	
	systemSignals := make(chan os.Signal, 1)

	//intercept external interreupt signals so we can turn off the server correctly
	signal.Notify(systemSignals, os.Interrupt)
	go func(){
	     <- systemSignals
	    
	    hub.Stop()
	    log.Println("Server stopped")
	    
	    os.Exit(2)
	}()

	err := hub.Run()

	if err != nil {
		log.Println(err);
	}
}

func printStats(stats *statbucket.StatBucket){
	log.Println("\n\n*** Execution statistics ***")
	log.Println(stats)
}