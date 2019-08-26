package server

import (
	"bufio"
	"fmt"
	"github.com/xesina/message-delivery/internal/message"
	"strconv"
	"strings"
)

const (
	receiveLogTpl  = "server: received %s message"
	responseLogTpl = "server: sent %s message response: %s"
)

// HandlerFunc is a message/command handler
type HandlerFunc func(*context) error

type context struct {
	id uint64
	rw *bufio.ReadWriter
}

func (server Server) handleUnknown(c *context) error {
	server.logger.Debugf(receiveLogTpl, "UNKNOWN")

	response := "UNKNOWN MESSAGE\n"
	_, err := c.rw.WriteString(response)
	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	server.logger.Debugf(responseLogTpl, "UNKNOWN", response[:len(response)-1])

	return nil
}

func (server Server) handleIdentity(c *context) error {
	server.logger.Debugf(receiveLogTpl, message.IdentityMsg)

	clientID := fmt.Sprintf("%s\n", strconv.FormatUint(c.id, 10))

	_, err := c.rw.WriteString(clientID)

	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	server.logger.Debugf(responseLogTpl, message.IdentityMsg, clientID[:len(clientID)-1])

	return nil
}

func (server Server) handleList(c *context) error {
	server.logger.Debugf(receiveLogTpl, message.ListMsg)

	ids := server.ListClientIDs()
	var response []string
	for _, id := range ids {
		if id == c.id {
			continue
		}
		response = append(response, fmt.Sprintf("%d", id))
	}

	rawResponse := strings.Join(response, ",")
	rawResponse = fmt.Sprintf("%s\n", rawResponse)

	_, err := c.rw.WriteString(rawResponse)
	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	server.logger.Debugf(responseLogTpl, message.ListMsg, rawResponse[:len(rawResponse)-1])

	return nil
}

func (server Server) handleSend(c *context) error {
	server.logger.Debugf(receiveLogTpl, message.SendMsg)

	m := message.Send{}
	err := m.Unmarshal(c.rw.Reader)
	if err != nil {
		return err
	}

	if len(m.Recipients) == 0 || len(m.Recipients) > 255 {
		_, err := c.rw.WriteString("ERR RECIPIENTS 1-255\n")
		if err != nil {
			return err
		}
		err = c.rw.Flush()
		if err != nil {
			return err
		}
		return nil
	}

	if len(m.Body) > 1<<20 {
		_, err := c.rw.WriteString("ERR TOO LARGE BODY 1M\n")
		if err != nil {
			return err
		}
		err = c.rw.Flush()
		if err != nil {
			return err
		}
		return nil
	}

	recipientsIDs := make(map[uint64]struct{})
	for _, id := range m.Recipients {
		recipientsIDs[id] = struct{}{}
	}

	server.cl.RLock()
	for cl, id := range server.clients {
		if _, ok := recipientsIDs[id]; !ok || id == c.id {
			continue
		}
		incoming := message.NewIncoming(c.id, m.Body)
		_, err := cl.Write(incoming.Marshal())
		if err != nil {
			return err
		}

	}
	server.cl.RUnlock()

	sendResponse := fmt.Sprintf("DONE\n")

	_, err = c.rw.WriteString(sendResponse)
	if err != nil {
		return err
	}

	err = c.rw.Flush()
	if err != nil {
		return err
	}
	server.logger.Debugf(responseLogTpl, message.SendMsg, sendResponse[:len(sendResponse)-1])

	return nil
}
