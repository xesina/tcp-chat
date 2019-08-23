package server

import (
	"log"
	"net"
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
