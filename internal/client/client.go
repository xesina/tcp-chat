package client

import (
	"bufio"
	"fmt"
	"github.com/xesina/message-delivery/internal/message"
	"io"
	"net"
	"strconv"
	"strings"
)

// IncomingMessage represents an incoming msg to a client
type IncomingMessage struct {
	SenderID uint64
	Body     []byte
}

// Client is implements request side of message protocol to easily connect
// and communicate with server
type Client struct {
	conn     net.Conn
	shutdown chan bool
}

// New creates and returns a new Client
func New() *Client {
	return &Client{
		shutdown: make(chan bool),
	}
}

// Connect will try to connect to the TCP server at the given address and port
func (c *Client) Connect(serverAddr *net.TCPAddr) error {
	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("client: connection failed: %s", err)
	}
	c.conn = conn

	return nil
}

// Close will terminates connection to the server
func (c *Client) Close() error {
	close(c.shutdown)
	c.conn.Close()
	return nil
}

// WhoAmI will sends a IDENTITY msg to server and returns the current
// client id
func (c *Client) WhoAmI() (uint64, error) {
	reader := bufio.NewReader(c.conn)
	msg := message.NewIdentity()
	_, err := c.conn.Write(msg.Marshal())
	if err != nil {
		return 0, fmt.Errorf("client: sending identity message failed: %s", err)
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	r := strings.TrimSpace(response)

	id, err := strconv.Atoi(r)
	if err != nil {
		return 0, fmt.Errorf("client: inavlid id received from the server: %s", err)
	}

	return uint64(id), nil
}

// ListClientIDs lists all clients connected to server
func (c *Client) ListClientIDs() ([]uint64, error) {
	reader := bufio.NewReader(c.conn)
	var ids []uint64
	msg := message.NewList()
	_, err := c.conn.Write(msg.Marshal())
	if err != nil {
		return ids, fmt.Errorf("client: sending list message failed: %s", err)
	}

	response, err := reader.ReadString('\n')
	r := strings.TrimSpace(response)

	if r == "" {
		return ids, nil
	}
	rr := strings.Split(r, ",")

	for _, id := range rr {
		i, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return ids, fmt.Errorf("client: inavlid id in list received from the server: %s", err)
		}
		ids = append(ids, i)
	}

	return ids, nil
}

// SendMsg sends a message using SEND msg with given ids and the payload
func (c *Client) SendMsg(recipients []uint64, body []byte) error {
	msg := message.NewSend(recipients, body)
	_, err := c.conn.Write(msg.Marshal())
	if err != nil {
		return fmt.Errorf("client: sending message failed: %s", err)
	}
	return nil
}

// HandleIncomingMessages once this method called this client switches to receive
// only mode and receives and decodes the INCOMING msg sent to the client and
// will forward the received msg to the given write-only channel.
func (c *Client) HandleIncomingMessages(writeCh chan<- IncomingMessage) {
	notify := make(chan error)
	r := bufio.NewReader(c.conn)

	for {
		select {
		case <-c.shutdown:
			close(notify)
			return
		case err := <-notify:
			fmt.Println("client: got an error:", err)

			if err == io.EOF {
				close(notify)
				c.Close()
				fmt.Println("client: connection dropped message", err)
				return
			}
		default:
			msg, err := r.ReadString('\n')
			if err != nil {
				notify <- err
				return
			}
			msg = strings.TrimSpace(msg)

			s, err := r.ReadString('\n')
			s = strings.TrimSpace(s)
			if err != nil {
				notify <- err
				continue
			}
			sender, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				notify <- err
				continue
			}
			body, err := r.ReadString('\n')
			if err != nil {
				notify <- err
				continue
			}
			body = strings.TrimSpace(body)
			body = fmt.Sprintf("%s", body)
			// TODO: parse sender-id and payload
			writeCh <- IncomingMessage{
				SenderID: sender,
				Body:     []byte(body),
			}
		}
	}
}
