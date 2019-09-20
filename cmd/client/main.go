package main

import (
	"bufio"
	"fmt"
	"github.com/xesina/tcp-chat/internal/client"
	"github.com/xesina/tcp-chat/internal/message"
	"net"
	"os"
	"strconv"
	"strings"
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

	//clientCh := make(chan client.IncomingMessage)
	//go cl.HandleIncomingMessages(clientCh)
	//go func() {
	//	for d := range clientCh {
	//		fmt.Printf("new message: sender: %d msg:%s\n", d.SenderID, string(d.Body))
	//	}
	//}()

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

	promptLoop(cl)

	cl.Close()
}

func promptLoop(cl *client.Client) {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		command, err := r.ReadString('\n')
		command = strings.ToUpper(strings.TrimSpace(command))
		if err != nil {
			panic(err)
		}

		switch command {
		case message.IdentityMsg:
			id, err := cl.WhoAmI()
			if err != nil {
				fmt.Println("WhoAmI message failed:", err)
				os.Exit(1)
			}

			fmt.Println("received id:", id)

		case message.ListMsg:
			ids, err := cl.ListClientIDs()
			if err != nil {
				fmt.Println("List message failed:", err)
				os.Exit(1)
			}

			fmt.Println("received ids:", ids)

		case message.SendMsg:
			line, err := r.ReadString('\n')
			line = strings.TrimSpace(line)
			if err != nil {
				panic(err)
			}
			rr := strings.Split(line, ",")
			var recipients []uint64
			for _, id := range rr {
				i, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					panic(err)
				}
				recipients = append(recipients, i)
			}

			body, err := r.ReadString('\n')
			body = strings.TrimSpace(body)
			if err != nil {
				panic(err)
			}

			err = cl.SendMsg(recipients, []byte(body))
			if err != nil {
				panic(err)
			}
		}

	}
}
