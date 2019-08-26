package message

// message package represents the protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	// IdentityMsg message name
	IdentityMsg = "IDENTITY"
	// ListMsg message name
	ListMsg = "LIST"
	// SendMsg message name
	SendMsg = "SEND"
	// IncomingMsg message name
	IncomingMsg = "INCOMING"
)

// ReadStringArg reads a string terminated in `newline` from `r`
// trims the `newline`
func ReadStringArg(r *bufio.Reader) (string, error) {
	arg, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	arg = strings.TrimSpace(arg)
	return arg, nil
}

// ReadBytesArg reads an array of bytes terminated in `newline` from `r`
// trims the `newline`
func ReadBytesArg(r *bufio.Reader) ([]byte, error) {
	arg, err := r.ReadBytes('\n')
	if err != nil {
		return []byte{}, err
	}
	arg = bytes.TrimSpace(arg)
	return arg, nil
}

// Read reads a message or command from r, trims it and convert it
// to uppercase and then return it
func Read(r *bufio.Reader) (string, error) {
	msg, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	msg = strings.ToUpper(strings.TrimSpace(msg))
	return msg, nil
}

// Identity represents an IDENTITY msg structure
type Identity struct{}

// NewIdentity creates a new instance of identity message
func NewIdentity() *Identity {
	return &Identity{}
}

// Marshal encodes the identity msg
func (m Identity) Marshal() []byte {
	return []byte(fmt.Sprintf("%s\n", IdentityMsg))
}

// List represents an LIST msg structure
type List struct{}

// NewList creates a new instance of list message
func NewList() *List {
	return &List{}
}

// Marshal encodes the list msg
func (m List) Marshal() []byte {
	return []byte(fmt.Sprintf("%s\n", ListMsg))
}

// Send represents an SEND msg structure
type Send struct {
	Recipients []uint64
	Body       []byte
}

// NewSend creates a new instance of send message
func NewSend(rr []uint64, b []byte) *Send {
	return &Send{
		Recipients: rr,
		Body:       b,
	}
}

// Marshal encodes the send msg
func (m Send) Marshal() []byte {
	rr := joinRecipients(m.Recipients)
	return []byte(fmt.Sprintf("%s\n%s\n%s\n", SendMsg, rr, string(m.Body)))
}

// Unmarshal decodes the send msg
func (m *Send) Unmarshal(r *bufio.Reader) error {
	s, err := ReadStringArg(r)
	if err != nil {
		return err
	}
	rr := strings.Split(s, ",")

	for _, recipientID := range rr {
		id, err := strconv.ParseUint(recipientID, 10, 64)
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

// Incoming represents an INCOMING msg structure
type Incoming struct {
	sender uint64
	body   []byte
}

// NewIncoming creates a new instance of incoming message
func NewIncoming(s uint64, b []byte) *Incoming {
	return &Incoming{
		sender: s,
		body:   b,
	}
}

// Marshal encodes the incoming msg
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
