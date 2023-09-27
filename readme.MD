# In Memory DB

## How does Listen System call works:
 - does listen system call loops infinitely over the TCP connection network buffer ?
   - The listen system call in the context of TCP (Transmission Control Protocol) does not loop infinitely over the TCP connection network buffer. Instead, the listen system call prepares a socket to accept incoming connections on a specific port, and it sets the socket's status to "listening." It doesn't directly deal with the network buffer.
     - Here's how it works:
        - **Setting up the Socket**: When a program calls the socket system call to create a socket, it specifies the type of socket (e.g., SOCK_STREAM for TCP) and other parameters. This socket is then associated with a local endpoint, typically an IP address and port.
        - **Binding:** After creating the socket, the program typically calls the bind system call to associate the socket with a specific local address and port. This is the address to which clients will connect.
        - **Listening**: Once the socket is bound, the program calls the listen system call. This sets up the socket to accept incoming connections on the specified local address and port. The socket transitions to a "listening" state.
        - **Accepting Connections:** After the socket is in the listening state, the program usually enters a loop, repeatedly calling the accept system call. The accept call blocks until a client initiates a connection.
        - **Accepting a Connection:** When a connection request arrives from a client, the accept system call returns a new socket. This new socket represents the established connection to that client.
        - **Handling Data:** The program can use this new socket to read data sent by the client or write data to the client. This is where data is transferred between the server and the client.
        - **Closing the Connection:** When the server is done with a particular client, it closes the socket associated with that client to terminate the connection. The server can continue to accept new connections and handle them in a similar manner.
 - The network buffer management, which involves reading and writing data to and from the network, is typically handled at a lower level by the networking stack of the operating system. The listen and accept system calls primarily deal with managing the status of the socket, accepting new connections, and providing a new socket for communication once a connection is established.
 - In summary, the listen system call doesn't loop infinitely over the TCP connection network buffer; rather, it prepares the socket to accept incoming connections, and the actual data transfer occurs through separate socket operations like read and write. The operating system's networking stack manages the network buffers to ensure data is sent and received reliably.

 - Command to send request to the server:
   - 