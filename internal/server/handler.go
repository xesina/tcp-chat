package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type HandleFunc func(net.Conn) error

func (server Server) handleUnknown(conn net.Conn) error {
	log.Print("Received UNKNOWN message")
	_, err := conn.Write([]byte("unknown message\n"))
	if err != nil {
		return err
	}
	return nil
}

func (server Server) handleIdentity(conn net.Conn) error {
	log.Print("Received IDENTITY message")

	server.cl.RLock()
	id := server.clients[conn]
	server.cl.RUnlock()

	clientId := fmt.Sprintf("%s\n", strconv.FormatUint(id, 10))

	_, err := conn.Write([]byte(clientId))
	if err != nil {
		return err
	}
	return nil
}
