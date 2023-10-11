package server

import (
	"github.com/inmemdb/inmem/config"
	"io"
	"log"
	"net"
	"strconv"
)

func RunInMemDBSyncServer() {
	log.Println("Running Sync TCP server on", config.Host, config.Port)
	var clients int

	//INFO: 1. Listen call:
	//This creates a new server socket and binds it with the network address (host:port) on which the server process keeps waiting
	//for a new tcp connection from a client
	srvListnerSockt, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	for {
		//INFO: 2: Accept call:
		//Accept creates a new socket for a new TCP client connection, it is a blocking call
		//and waits until the client has successfully established the TCP connection with the server
		// When a successful connection is established it returnes a new connection (socket) for the
		//TCP client
		newClientConn, err := srvListnerSockt.Accept()
		if err != nil {
			panic(err)
		}

		clients += 1
		log.Println("New Client is Connected to our server with address: ", newClientConn.RemoteAddr(), "Client Numer=", clients)
		handleConnection(newClientConn, clients)
		clients -= 1
	}
}

func handleConnection(newConnSocket net.Conn, clients int) error {
	defer newConnSocket.Close()
	//Process on the connection that is established, continuously loop over the connection to keep reading
	//whatever is sent by the client over the TCP connection.
	for {
		req, err := readCommand(newConnSocket)
		if err != nil {

			log.Println("Client disconnected", newConnSocket.RemoteAddr(), " ClientId = ", clients)
			clients -= 1

			//If client sends and EOF in case of graceful termination,
			//we simply break out of the loop and terminate the connection.
			if err == io.EOF {
				break
			}
			log.Println("Read Connection Error: ", err)
			break
		}
		log.Println("Req Sent is: ", req)
		if err = respond(req, newConnSocket); err != nil {
			log.Println("error responding to client: ", err)
		}
	}
	return nil
}

func readCommand(c net.Conn) (*Command, error) {
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
	cmdTokens, err := DecodeInputCommand(buf)
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

func respond(req *Command, c net.Conn) error {
	data, err := req.EvalCommand()
	if err != nil {
		c.Write(EncodeError(err))
	}
	_, err = c.Write(data)
	if err != nil {
		c.Write(EncodeError(err))
	}
	return nil
}
