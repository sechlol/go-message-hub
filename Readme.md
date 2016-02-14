# Message Delivery System

## Installation
Be sure that all the project files are under the directory **$GOPATH/src/github.com/sech90/go-message-hub**. Run the installer script in the project folder to install the dependencies and make the executables to run the simulation. After, there should be two executables: **start_server** and **start_simulation**

```sh install.sh```

## Test
Run the do_test script to run all the tests and benchmarks

```sh do_test.sh```

## Simulation
Firstly run the **start_server** executable. Use the -stat flag to show the collected statistics before closing. Use -port=xxxx to specify a port number (default 9999). Then run the **start_simulation** executable to simulate a message exchange between clients. Every client will send a Relay Message *nmex* times, the recipients will be all the other *ncli*-1 clients. Here are all the flags with their defaults:

* -addr="localhost"
* -port=9999
* -size=10240 (set size for message payload in bytes)
* -ncli=100 (numbers of clients to connect)
* -nmex=100 (number of times one client will broadcast the payload)
* -i=1 (a client's time delay between send requests, in milliseconds)
* -stat=true (show progress and statistics)

For a very light simulation, these may be good values:
``` 
./start_simulation -size=1024 -ncli=2 -nmex=10
./start_simulation -size=1024 -ncli=10 -nmex=20
```

For heavy simulations, set the maximum values for the payload and connected users, and change the number of message for user
``` 
./start_simulation -size=1024000 -ncli=255 -nmex=20
```

## Description
The application is composed in two main parts: a **Hub** and a **Client**. 

### Hub
The Hub acts like a server, listening to one port and accepting all the new connected clients in a loop. When a client connects, starts a *goroutine* to process the connection: 

* Get a unique ID from **IdPool**
* Create a **MexSocket** object and start it in "Server mode"
* Save it in a thread-safe map

After the initialization, the goroutine enters in a for loop, constantly reading incoming Requests from the client (mediated via the MexSocket). When a new request is received, the Hub starts a new goroutine to handle the operation. The for loop will end only when the MexSocket terminates by closing its quit channel

### Client
Very similar to the Hub, the Client connects to a given address and starts a goroutine in which:

* Starts a **MexSocket** in "Client mode" to handle the communication over the network. 
* Enters a for loop listening to MexSocket channels
* If a new answer is received, progagate it to the outside
* If MexSocket terminates, exit the loop

After the main goroutine is started, the clients sends an ID request and waits until it receives an answer. After this, the client is correctly connected and ready to use
