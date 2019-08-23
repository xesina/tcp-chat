package message

// message package represents the protocol

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	IdentityMsg = "IDENTITY"
	ListMsg     = "LIST"
)

func Read(r io.Reader) (string, error) {
	msg, err := bufio.NewReader(r).ReadString('\n')
	if err != nil {
		return "", err
	}

	msg = strings.ToUpper(strings.TrimSpace(msg))

	return msg, nil
}

type Identity struct{}

func NewIdentity() *Identity {
	return &Identity{}
}

func (m Identity) Marshal() []byte {
	return []byte(fmt.Sprintf("%s\n", IdentityMsg))
}

type List struct{}

func NewList() *List {
	return &List{}
}

func (m List) Marshal() []byte {
	return []byte(fmt.Sprintf("%s\n", ListMsg))
}
