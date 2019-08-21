package main

import (
	"fmt"
	"github.com/xesina/message-delivery/internal/server"
	"net"
	"os"
)

const serverPort = 50000

func main() {
	srv := server.New()
	tcpAddr := net.TCPAddr{Port: serverPort}

	err := srv.Start(&tcpAddr)
	if err != nil {
		fmt.Println("starting server failed: ", err)
		os.Exit(1)
	}
}
