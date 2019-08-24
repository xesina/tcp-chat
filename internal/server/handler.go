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

	clientId := fmt.Sprintf("%s\n%s\n", message.IdentityMsgResponse, strconv.FormatUint(c.id, 10))

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
	rawResponse = fmt.Sprintf("%s\n%s\n", message.ListMsgResponse, rawResponse)

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
