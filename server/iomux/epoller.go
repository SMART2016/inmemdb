package iomux

import (
	"errors"
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/inmemdb/inmem/config"
)

// Epoll is a epoll based poller.
type Epoll struct {
	//The Epoller instance FD, file descriptor
	//referring to the kqueue instance.
	//Events will be monitored on this kqueue.
	epollerFd int

	//The kqueue events container for a client connection
	changes []syscall.Kevent_t

	//The kqueue events container for all
	//events that can occur on a client connection
	events []syscall.Kevent_t

	//The client connections that will be stored in the map
	conns map[int]net.Conn

	mu sync.RWMutex
}

// NewPoller creates a new poller instance.
func NewPoller(serverFd int) (*Epoll, error) {
	// Here we are assuming there will be only 128 events per client connection
	return NewPollerWithBuffer(serverFd, 128)
}

func NewPollerWithBuffer(serverFd int, count int) (*Epoll, error) {

	//Create an EPOLLER kqueue instance ,
	//which will be used to queue client coneections and there events
	p, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}

	//Adding a user-defined event filter to the epoll instance
	//is a common practice in event-driven programming to
	//ensure that the epoll instance doesn't block indefinitely
	//when waiting for events.
	_, err = syscall.Kevent(p, []syscall.Kevent_t{{
		//FD for the socket on which to listen for the user event
		//0 means the standard input file descriptor.
		Ident: uint64(serverFd),
		//READ and WRITE filters to be monitored on the file descriptor
		Filter: syscall.EVFILT_READ | syscall.EVFILT_WRITE,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		panic(err)
	}

	return &Epoll{
		epollerFd: p,
		conns:     make(map[int]net.Conn),
		events:    make([]syscall.Kevent_t, count),
	}, nil
}

func (e *Epoll) Add(conn net.Conn) error {
	//1. fetch the FD for the connection
	fd := socketFD(conn)

	//2. Check if FD could be set to non blocking state
	//After setting the file descriptor to non-blocking mode,
	//any I/O operations (like syscall.Read or syscall.Write) on that file descriptor will return immediately,
	//even if there is no data available or if the operation can't be completed immediately.
	if e := syscall.SetNonblock(int(fd), true); e != nil {
		return errors.New("unix.SetNonblock failed")
	}

	//3. Create a kqueue event to monitor any IO on the FD above and add it to the change list for events
	//changes list is holding all client fd's for whome we want to monitor IO
	e.mu.Lock()
	defer e.mu.Unlock()
	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_EOF, Filter: syscall.EVFILT_READ,
		},
	)

	//4. Finally add the current client connection to the connection map
	e.conns[fd] = conn

	return nil
}

func (e *Epoll) Close(closeConns bool) error {
	if closeConns {
		for _, conn := range e.conns {
			conn.Close()
		}
	}

	e.conns = nil
	e.changes = nil

	return syscall.Close(e.epollerFd)
}

func (e *Epoll) Wait() ([]net.Conn, error) {

	//1. Waits for any IO events on the FD's added in the changes List
	n, err := syscall.Kevent(e.epollerFd, e.changes, e.events, nil)
	if err != nil {
		return nil, err
	}

	//2. Fetches the connections from the events list who all have data to
	//read and adds them to the list of connections which needs to be processed
	e.mu.RLock()
	defer e.mu.RUnlock()
	conns := make([]net.Conn, 0, n)
	for i := 0; i < n; i++ {
		conn := e.conns[int(e.events[i].Ident)]
		//Close the connections for whome EOF has been sent
		if (e.events[i].Flags & syscall.EV_EOF) == syscall.EV_EOF {
			conn.Close()
		}
		conns = append(conns, conn)
	}
	return conns, nil
}

// Used to wait for the passed connection only.
func (e *Epoll) WaitForCurrentConn(c net.Conn) ([]net.Conn, error) {

	//1. Waits for any IO events on the FD's added in the changes List
	n, err := syscall.Kevent(e.epollerFd, e.changes, e.events, nil)
	if err != nil {
		return nil, err
	}

	//2. Fetches the connections from the events list who all have data to
	//read and adds them to the list of connections which needs to be processed
	e.mu.RLock()
	defer e.mu.RUnlock()
	conns := make([]net.Conn, 0, n)
	for i := 0; i < n; i++ {
		// Finds events only for current passed connection and processes only for current connection
		if int(e.events[i].Ident) == socketFD(c) {
			conn := e.conns[int(e.events[i].Ident)]
			//Close the connections for whome EOF has been sent
			if (e.events[i].Flags & syscall.EV_EOF) == syscall.EV_EOF {
				conn.Close()
			}
			conns = append(conns, conn)
		}
	}
	return conns, nil
}

// Remove removes a connection from the poller.
// If close is true, the connection will be closed.
func (e *Epoll) Remove(conn net.Conn) error {
	defer conn.Close()
	fd := socketFD(conn)
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.changes) <= 1 {
		e.changes = nil
	} else {
		changes := make([]syscall.Kevent_t, 0, len(e.changes)-1)
		ident := uint64(fd)
		for _, ke := range e.changes {
			if ke.Ident != ident {
				changes = append(changes, ke)
			}
		}
		e.changes = changes
	}

	delete(e.conns, fd)

	return nil
}
func (e *Epoll) bind(serverSocketFd int, ipBytes [4]byte) {
	err := syscall.Bind(serverSocketFd, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: ipBytes,
	})
	if err != nil {
		log.Println("Failed to Bind server socket ", err)
		panic(err)
	}
}
