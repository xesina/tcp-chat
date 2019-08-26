package server

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xesina/message-delivery/internal/message"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Server implements a TCP message server
type Server struct {
	debug    bool
	logger   *logrus.Logger
	listener *net.TCPListener

	id      uint64
	cl      *sync.RWMutex
	clients map[net.Conn]uint64

	hl      *sync.RWMutex
	handler map[string]HandlerFunc

	shutdown chan bool
}

// New creates and sets up a new server instance
func New(debug bool) *Server {
	s := &Server{
		debug:    debug,
		logger:   logrus.New(),
		clients:  make(map[net.Conn]uint64),
		handler:  make(map[string]HandlerFunc),
		cl:       &sync.RWMutex{},
		hl:       &sync.RWMutex{},
		shutdown: make(chan bool),
	}

	s.logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	if s.debug {
		s.logger.SetLevel(logrus.DebugLevel)
	}

	s.registerHandlers()

	return s
}

func (server *Server) registerHandlers() {
	server.HandleFunc(message.IdentityMsg, server.handleIdentity)
	server.HandleFunc(message.ListMsg, server.handleList)
	server.HandleFunc(message.SendMsg, server.handleSend)
}

// Start will bootstrap and starts the server and connection handling
// loop.
func (server *Server) Start(laddr *net.TCPAddr) error {
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return fmt.Errorf("error listening: %s", err)
	}
	server.listener = l
	defer l.Close()

	fmt.Printf("Listening on %s:%d\n", laddr.IP, laddr.Port)
	if server.debug {
		fmt.Println("Server is running on debug mode.")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			server.logger.Errorf("server: failed accepting a connection request: %s", err)
			continue
		}
		server.logger.Debug("server: handle incoming connection")
		go server.handleConnection(conn)
	}
}

// ListClientIDs returns the current active clients ids
func (server *Server) ListClientIDs() []uint64 {
	var ids []uint64
	server.cl.RLock()
	defer server.cl.RUnlock()
	for _, id := range server.clients {
		ids = append(ids, id)
	}
	return ids
}

// Stop Stops accepting connections and close the existing ones
func (server *Server) Stop() error {
	fmt.Println("Stop accepting connections and close the existing ones")
	server.shutdown <- true
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

// HandleFunc registers a new message with associated handler
func (server *Server) HandleFunc(name string, f HandlerFunc) {
	server.hl.Lock()
	server.handler[name] = f
	server.hl.Unlock()
}

// HandleMessage finds the appropriate handler based on msg and passes control to it
func (server *Server) HandleMessage(name string, ctx *context) error {
	h, ok := server.Handler(name)
	if !ok {
		h = server.handleUnknown
	}
	return h(ctx)
}

// Handler returns the Handler associated with a msg
func (server *Server) Handler(msg string) (HandlerFunc, bool) {
	server.hl.RLock()
	h, ok := server.handler[msg]
	server.hl.RUnlock()
	return h, ok
}

// ClientID returns the unsigned integer  associated with client connection
func (server *Server) ClientID(conn net.Conn) uint64 {
	server.cl.RLock()
	defer server.cl.RUnlock()
	return server.clients[conn]
}

func (server *Server) handleConnection(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	notify := make(chan error)
	server.registerClient(conn)
	go func() {
		for {
			msg, err := message.Read(rw.Reader)
			if err != nil {
				notify <- err
				return
			}
			ctx := &context{
				id: server.ClientID(conn),
				rw: rw,
			}
			err = server.HandleMessage(msg, ctx)
			if err != nil {
				notify <- err
			}
		}
	}()

	for {
		select {
		case <-server.shutdown:
			break
		case err := <-notify:
			server.logger.Debug("server: an error occurred: ", err)
			if err == io.EOF {
				server.logger.Debug("server: closing connection because: ", err)
				server.deregisterClient(conn)
				return
			}
		}
	}
}
