package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"syscall"

	"github.com/inmemdb/inmem/config"
	iomux "github.com/inmemdb/inmem/server/iomux"
	"github.com/inmemdb/inmem/server/response"
)

type Poller interface {
	Add(conn net.Conn) error
	Close(closeConns bool) error
	//Wait() ([]net.Conn, error)
	WaitForCurrentConn(conn net.Conn) ([]net.Conn, error)
	Remove(conn net.Conn) error
}

type AsyncServer struct {
	poller      Poller
	listener    net.Listener
	threadCount int
}

func NewAsyncServer() *AsyncServer {
	//1. Create a Server Socket and fetch the FD for the server
	serverSocketFd, listener := createServerSocket()
	defer syscall.Close(serverSocketFd)

	//2. Create an EPOLL instance
	epoller, err := iomux.NewPoller(serverSocketFd)
	if err != nil {
		log.Println("Error Creating an Epoller instance, Err: ", err)
	}

	return &AsyncServer{
		epoller, listener, 0,
	}

}
func (as *AsyncServer) RunInMemDBASyncServer() {
	log.Println("Running Async TCP server on", config.Host, config.Port)
	//Configuration to limit the number of concurrent Clients at a time
	//maxClients := 20000

	for {
		//A. Start Accepting for conections
		conn, err := as.listener.Accept()
		if err != nil {
			log.Println("Error in establishing Client Connection, Err: ", err)
		}

		//B. use a goroutene to process all connections as and when data is available to process.
		//Creates a goroutene for each connection and remains for all connectection
		//go as.poll()
		go as.pollForConn(conn)

		//c. Add the connection to the EPOLL event container in EPoll instance for monitoring
		err = as.poller.Add(conn)
	}
}

// Used to poll only for current connection not all connections.
func (as *AsyncServer) pollForConn(c net.Conn) {
	for {
		conns, err := as.poller.WaitForCurrentConn(c)
		if err != nil {
			continue
		}

		for _, conn := range conns {
			as.handleAsyncConnection(conn)
		}
	}
}

// func (as *AsyncServer) poll() {
// 	for {
// 		conns, err := as.poller.Wait()
// 		if err != nil {
// 			continue
// 		}

// 		for _, conn := range conns {
// 			as.handleAsyncConnection(conn)
// 		}
// 	}
// }

func (as *AsyncServer) handleAsyncConnection(clientConn net.Conn) error {
	//Process on the connection that is established, continuously loop over the connection to keep reading
	//whatever is sent by the client over the TCP connection.
	for {
		req, err := readAsyncClientCommand(clientConn)
		if err != nil {
			as.poller.Remove(clientConn)
			log.Println("Read Connection Error: ", err)
			break
		}
		log.Println("Req Sent is: ", req)
		if err = respondAsyncClient(req, clientConn); err != nil {
			log.Println("error responding to client: ", err)
		}
	}
	return nil
}

func readAsyncClientCommand(c net.Conn) (*Command, error) {
	//INFO: Max reads is 512 bytes on a single read call ,
	//for larger inputs we keep calling readCommand until
	//we recieve EOF from client
	var buf []byte = make([]byte, 512)

	//INFO: This is a blocking call and blocks until the client
	//sends some bytes to the server over the TCP connection.
	_, err := c.Read(buf)
	if err != nil {
		return nil, err
	}

	//INFO: The requests or commands are submitted to The server as array of strings encoded in RESP,
	//So the command is decoded into simple array
	cmdTokens, err := response.DecodeInputCommand(buf)
	if err != nil {
		return nil, err
	}

	//INFO: we specifically mentioned the end of the buffer slice ,
	//otherwise if the buffer is not full it can have garbage
	return &Command{
		Cmd:  cmdTokens[0],
		Args: cmdTokens[1:],
	}, nil
}

func respondAsyncClient(req *Command, c net.Conn) error {
	data, err := req.EvalCommand()
	if err != nil {
		c.Write(response.EncodeError(err))
	}
	_, err = c.Write(data)
	if err != nil {
		c.Write(response.EncodeError(err))
	}
	return nil
}

func createServerSocket() (int, net.Listener) {
	//serverSocketFd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	//if err != nil {
	//	log.Println("Failed to create server socket ", err)
	//	panic(err)
	//}

	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	// Get the underlying file descriptor
	file, err := listener.(*net.TCPListener).File()
	if err != nil {
		fmt.Println("Error getting file descriptor:", err)
		return 0, nil
	}

	// Get the integer file descriptor
	serverSocketFd := int(file.Fd())
	return serverSocketFd, listener
}
