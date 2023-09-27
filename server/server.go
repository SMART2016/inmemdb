package server

import (
	"github.com/inmemdb/inmem/config"
	"io"
	"log"
	"net"
	"strconv"
)

func RunInMemDBServer() {
	log.Println("Running TCP server on", config.Host, config.Port)
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
		newConnSocket, err := srvListnerSockt.Accept()
		if err != nil {
			panic(err)
		}

		clients += 1
		log.Println("New Client is Connected to our server with address: ", newConnSocket.RemoteAddr(), "Client Numer=", clients)
		handleConnection(newConnSocket, clients)
	}
}

func handleConnection(newConnSocket net.Conn, clients int) error {
	defer newConnSocket.Close()
	//Process on the connection that is established, continuously loop over the connection to keep reading
	//whatever is sent by the client over the TCP connection.
	for {
		req, err := readReq(newConnSocket)
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

func readReq(c net.Conn) (string, error) {
	//INFO: Max reads is 512 bytes on a single read call ,
	//for larger inputs we keep calling readReq until
	//we recieve EOF from client
	var buf []byte = make([]byte, 512)

	//INFO: This is a blocking call and blocks until the client
	//sends some bytes to the server over the TCP connection.
	n, err := c.Read(buf)
	if err != nil {
		return "", err
	}

	//INFO: we specifically mentioned the end of the buffer slice ,
	//otherwise if the buffer is not full it can have garbage
	return string(buf[:n]), nil
}

func respond(req string, c net.Conn) error {
	if _, err := c.Write([]byte(req)); err != nil {
		return err
	}
	return nil
}
