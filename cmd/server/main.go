package main

import (
	"flag"
	"fmt"
	"github.com/xesina/tcp-chat/internal/server"
	"net"
	"os"
)

func main() {
	var (
		port  int
		debug bool
	)

	flag.IntVar(&port, "port", 50000, "Server port")
	flag.BoolVar(&debug, "debug", false, "Debug mode")

	flag.Parse()

	srv := server.New(debug)
	tcpAddr := net.TCPAddr{Port: port}

	err := srv.Start(&tcpAddr)
	if err != nil {
		fmt.Println("starting server failed: ", err)
		os.Exit(1)
	}
}
