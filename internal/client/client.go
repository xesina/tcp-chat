package client

import (
	"fmt"
	"net"
)

type IncomingMessage struct {
	SenderID uint64
	Body     []byte
}

type Client struct {
	conn net.Conn
}

func New() *Client {
	return &Client{}
}

func (c *Client) Connect(serverAddr *net.TCPAddr) error {
	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("client: connection failed: %s", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	fmt.Println("Close the connection to the server")
	c.conn.Close()
	return nil
}

func (c *Client) WhoAmI() (uint64, error) {
	fmt.Println("TODO: Fetch the ID from the server")
	return 0, nil
}

func (c *Client) ListClientIDs() ([]uint64, error) {
	fmt.Println("TODO: Fetch the IDs from the server")
	return nil, nil
}

func (c *Client) SendMsg(recipients []uint64, body []byte) error {
	fmt.Println("TODO: Send the message to the server")
	return nil
}

func (c *Client) HandleIncomingMessages(writeCh chan<- IncomingMessage) {
	fmt.Println("TODO: Handle the messages from the server")
}
