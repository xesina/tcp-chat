package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	listener *net.TCPListener

	id      uint64
	cl      sync.RWMutex
	clients map[net.Conn]uint64

	hl      sync.RWMutex
	handler map[string]HandleFunc
}

func New() *Server {
	s := &Server{
		clients: make(map[net.Conn]uint64),
		handler: make(map[string]HandleFunc),
	}

	s.HandleFunc("IDENTITY", s.handleIdentity)
	return s
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
	server.cl.Lock()
	server.clients[c] = atomic.AddUint64(&server.id, 1)
	server.cl.Unlock()
}

func (server *Server) deregisterClient(conn net.Conn) {
	server.cl.Lock()
	delete(server.clients, conn)
	server.cl.Unlock()
}

func (server *Server) HandleFunc(name string, f HandleFunc) {
	server.hl.Lock()
	server.handler[name] = f
	server.hl.Unlock()
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	notify := make(chan error)

	fmt.Println("new Connection received")
	server.registerClient(conn)

	go func() {
		for {
			msg, err := bufio.NewReader(conn).ReadString('\n')
			msg = strings.ToUpper(strings.TrimSpace(msg))

			if err != nil {
				notify <- err
				return
			}

			server.hl.RLock()
			handleCommand, ok := server.handler[msg]
			server.hl.RUnlock()
			if !ok {
				handleCommand = server.handleUnknown
			}
			err = handleCommand(conn)
			if err != nil {
				notify <- err
			}

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
			server.cl.RLock()
			fmt.Printf("connection id: %d still alive\n", server.clients[conn])
			server.cl.RUnlock()
		}
	}
}
