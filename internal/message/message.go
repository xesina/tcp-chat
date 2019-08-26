package message

// message package represents the protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	IdentityMsg = "IDENTITY"
	ListMsg     = "LIST"
	SendMsg     = "SEND"
	IncomingMsg = "INCOMING"
)

func ReadStringArg(r *bufio.Reader) (string, error) {
	arg, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	arg = strings.TrimSpace(arg)
	return arg, nil
}

func ReadBytesArg(r *bufio.Reader) ([]byte, error) {
	arg, err := r.ReadBytes('\n')
	if err != nil {
		return []byte{}, err
	}
	arg = bytes.TrimSpace(arg)
	return arg, nil
}

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

type Send struct {
	Recipients []uint64
	Body       []byte
}

func NewSend(rr []uint64, b []byte) *Send {
	return &Send{
		Recipients: rr,
		Body:       b,
	}
}

func (m Send) Marshal() []byte {
	rr := joinRecipients(m.Recipients)
	return []byte(fmt.Sprintf("%s\n%s\n%s\n", SendMsg, rr, string(m.Body)))
}

func (m *Send) Unmarshal(r *bufio.Reader) error {
	s, err := ReadStringArg(r)
	if err != nil {
		return err
	}
	rr := strings.Split(s, ",")

	for _, recipientId := range rr {
		id, err := strconv.ParseUint(recipientId, 10, 64)
		if err != nil {
			return err
		}
		m.Recipients = append(m.Recipients, id)
	}

	m.Body, err = ReadBytesArg(r)
	if err != nil {
		return err
	}
	return nil
}

type Incoming struct {
	sender uint64
	body   []byte
}

func NewIncoming(s uint64, b []byte) *Incoming {
	return &Incoming{
		sender: s,
		body:   b,
	}
}

func (m Incoming) Marshal() []byte {
	s := strconv.FormatUint(m.sender, 10)
	return []byte(fmt.Sprintf("%s\n%s\n%s\n", IncomingMsg, s, string(m.body)))
}

func joinRecipients(rr []uint64) string {
	var jr []string
	for _, id := range rr {
		jr = append(jr, strconv.FormatUint(id, 10))
	}
	return strings.Join(jr, ",")
}
