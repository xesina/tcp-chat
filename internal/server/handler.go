package server

import (
	"bufio"
	"fmt"
	"github.com/xesina/message-delivery/internal/message"
	"log"
	"strconv"
	"strings"
)

type HandleFunc func(*context) error

type context struct {
	id uint64
	rw *bufio.ReadWriter
}

func (server Server) handleUnknown(c *context) error {
	log.Print("Received UNKNOWN message")
	_, err := c.rw.WriteString("unknown message\n")
	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (server Server) handleIdentity(c *context) error {
	log.Print("Received IDENTITY message")

	clientId := fmt.Sprintf("%s\n", strconv.FormatUint(c.id, 10))

	_, err := c.rw.WriteString(clientId)

	if err != nil {
		return err
	}
	err = c.rw.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (server Server) handleList(c *context) error {
	log.Print("Received LIST message")

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
	return nil
}

func (server Server) handleSend(c *context) error {
	log.Print("Received Send message")
	m := message.Send{}
	err := m.Unmarshal(c.rw.Reader)
	if err != nil {
		return err
	}
	if len(m.Recipients) == 0 {
		_, err := c.rw.WriteString("invalid recipients")
		if err != nil {
			return err
		}
	}

	recipientsMap := make(map[uint64]struct{})
	for _, id := range m.Recipients {
		recipientsMap[id] = struct{}{}
	}

	server.cl.RLock()
	for cl, id := range server.clients {
		if _, ok := recipientsMap[id]; !ok || id == c.id {
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
	return nil
}
