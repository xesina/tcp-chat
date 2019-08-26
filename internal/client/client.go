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

type IncomingMessage struct {
	SenderID uint64
	Body     []byte
}

type Client struct {
	conn     net.Conn
	shutdown chan bool
}

func New() *Client {
	return &Client{
		shutdown: make(chan bool),
	}
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
	close(c.shutdown)
	c.conn.Close()
	return nil
}

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

func (c *Client) SendMsg(recipients []uint64, body []byte) error {
	msg := message.NewSend(recipients, body)
	_, err := c.conn.Write(msg.Marshal())
	if err != nil {
		return fmt.Errorf("client: sending message failed: %s", err)
	}
	return nil
}

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
