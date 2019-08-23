package main

import (
	"fmt"
	"github.com/xesina/message-delivery/internal/client"
	"net"
	"os"
	"time"
)

const serverPort = 50000

func main() {
	cl := client.New()
	tcpAddr := net.TCPAddr{Port: serverPort}

	err := cl.Connect(&tcpAddr)
	if err != nil {
		fmt.Println("connection to server failed: ", err)
		os.Exit(1)
	}

	id, err := cl.WhoAmI()
	if err != nil {
		fmt.Println("WhoAmI message failed:", err)
		os.Exit(1)
	}

	fmt.Println("received id:", id)

	ids, err := cl.ListClientIDs()
	if err != nil {
		fmt.Println("List message failed:", err)
		os.Exit(1)
	}

	fmt.Println("received ids:", ids)

	time.Sleep(time.Second * 10)
	cl.Close()
}
