package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
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
	fmt.Println("TODO: Stop accepting connections and close the existing ones")
	return nil
}

func (server *Server) assignId(c net.Conn) {
	l := len(server.clients)
	server.clients[c] = uint64(l + 1)
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	notify := make(chan error)

	fmt.Println("new Connection received")
	server.assignId(conn)
	fmt.Printf("%+v\n", server.clients)

	go func() {
		for {
			payload, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				notify <- err
				return
			}
			// basic echo server
			conn.Write([]byte(payload))
		}
	}()

	for {
		select {
		case err := <-notify:
			delete(server.clients, conn)
			if err == io.EOF {
				fmt.Println("connection dropped message", err)
				return
			}

		case <-time.After(time.Second * 20):
			fmt.Printf("connection id: %d still alive\n", server.clients[conn])
		}
	}
}
