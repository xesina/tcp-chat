package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	listener *net.TCPListener

	id uint64
	sync.RWMutex
	clients map[net.Conn]uint64
}

func New() *Server {
	return &Server{
		clients: make(map[net.Conn]uint64),
	}
}

func (server *Server) Start(laddr *net.TCPAddr) error {
	fmt.Println("Start handling client connections and messages")

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return fmt.Errorf("error listening: %s", err)
	}
	server.listener = l
	defer l.Close()

	fmt.Printf("Listening on %s:%d\n", laddr.IP, laddr.Port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go server.handleConnection(conn)
	}
	return nil
}

func (server *Server) ListClientIDs() []uint64 {
	fmt.Println("TODO: Return the IDs of the connected clients")
	return []uint64{}
}

func (server *Server) Stop() error {
	fmt.Println("Stop accepting connections and close the existing ones")
	return server.listener.Close()
}

func (server *Server) registerClient(c net.Conn) {
	server.Lock()
	server.clients[c] = atomic.AddUint64(&server.id, 1)
	server.Unlock()
}

func (server *Server) deregisterClient(conn net.Conn) {
	server.Lock()
	delete(server.clients, conn)
	server.Unlock()
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	notify := make(chan error)

	fmt.Println("new Connection received")
	server.registerClient(conn)

	go func() {
		for {
			payload, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				notify <- err
			}
			// basic echo server
			conn.Write([]byte(payload))
		}
	}()

	for {
		select {
		case err := <-notify:
			fmt.Println("got an error", err)

			if err == io.EOF {
				server.deregisterClient(conn)
				fmt.Println("connection dropped message", err)
				return
			}

		case <-time.After(time.Second * 20):
			server.RLock()
			fmt.Printf("connection id: %d still alive\n", server.clients[conn])
			server.RUnlock()
		}
	}
}
