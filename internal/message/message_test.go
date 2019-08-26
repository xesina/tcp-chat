package message

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestRead(t *testing.T) {
	tt := []struct {
		given *bufio.Reader
		want  string
		err   error
	}{
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("identity\n"))),
			want:  IdentityMsg,
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("IDENTITY\n"))),
			want:  IdentityMsg,
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("TEST"))),
			want:  "",
			err:   io.EOF,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("\n"))),
			want:  "",
			err:   nil,
		},
	}

	for _, tc := range tt {
		msg, err := Read(tc.given)
		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.want, msg)
	}
}

func TestReadStringArg(t *testing.T) {
	tt := []struct {
		given *bufio.Reader
		want  string
		err   error
	}{
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("arg1\n"))),
			want:  "arg1",
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("ARG1\n"))),
			want:  "ARG1",
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("ARG1    \n"))),
			want:  "ARG1",
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("1,2,3,4,5\n"))),
			want:  "1,2,3,4,5",
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("TEST"))),
			want:  "",
			err:   io.EOF,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("\n"))),
			want:  "",
			err:   nil,
		},
	}

	for _, tc := range tt {
		msg, err := ReadStringArg(tc.given)
		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.want, msg)
	}
}

func TestReadBytesArg(t *testing.T) {
	tt := []struct {
		given *bufio.Reader
		want  []byte
		err   error
	}{
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("arg1\n"))),
			want:  []byte("arg1"),
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("ARG1\n"))),
			want:  []byte("ARG1"),
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("ARG1    \n"))),
			want:  []byte("ARG1"),
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("1,2,3,4,5\n"))),
			want:  []byte("1,2,3,4,5"),
			err:   nil,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("TEST"))),
			want:  []byte(""),
			err:   io.EOF,
		},
		{
			given: bufio.NewReader(bytes.NewBuffer([]byte("\n"))),
			want:  []byte(nil),
			err:   nil,
		},
	}

	for _, tc := range tt {
		msg, err := ReadBytesArg(tc.given)
		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.want, msg)
	}
}

func TestIdentity_Marshal(t *testing.T) {
	expected := []byte("IDENTITY\n")
	msg := NewIdentity()
	actual := msg.Marshal()
	assert.Equal(t, expected, actual)
}

func TestList_Marshal(t *testing.T) {
	expected := []byte("LIST\n")
	msg := NewList()
	actual := msg.Marshal()
	assert.Equal(t, expected, actual)
}

func TestSend_Marshal(t *testing.T) {
	tt := []struct {
		recipients []uint64
		body       []byte
		want       []byte
	}{
		{
			recipients: []uint64{1, 2},
			body:       []byte("Hello"),
			want:       []byte("SEND\n1,2\nHello\n"),
		},
		{
			recipients: []uint64{},
			body:       []byte("Hello"),
			want:       []byte("SEND\n\nHello\n"),
		},
		{
			recipients: []uint64{1},
			body:       []byte("Hello\n"),
			want:       []byte("SEND\n1\nHello\n\n"),
		},
		{
			recipients: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			body:       []byte("Hello Friend!"),
			want:       []byte("SEND\n1,2,3,4,5,6,7,8,9,10\nHello Friend!\n"),
		},
	}

	for _, tc := range tt {
		msg := NewSend(tc.recipients, tc.body)
		actual := msg.Marshal()
		assert.Equal(t, tc.want, actual)
	}
}

func TestSend_Unmarshal(t *testing.T) {
	tt := []struct {
		given      *bufio.Reader
		recipients []uint64
		body       []byte
		hasErr     bool
	}{
		{
			given:      bufio.NewReader(bytes.NewBuffer([]byte("1,2\nHi\n"))),
			recipients: []uint64{1, 2},
			body:       []byte("Hi"),
			hasErr:     false,
		},
		{
			given:      bufio.NewReader(bytes.NewBuffer([]byte("1\nHi\n"))),
			recipients: []uint64{1},
			body:       []byte("Hi"),
			hasErr:     false,
		},
		{
			given:      bufio.NewReader(bytes.NewBuffer([]byte("\nHi\n"))),
			recipients: []uint64(nil),
			body:       []byte(nil),
			hasErr:     true,
		},
		{
			given:      bufio.NewReader(bytes.NewBuffer([]byte("1,2,3,4\n\n"))),
			recipients: []uint64{1, 2, 3, 4},
			body:       []byte(nil),
			hasErr:     false,
		},
		{
			given:      bufio.NewReader(bytes.NewBuffer([]byte(""))),
			recipients: []uint64(nil),
			body:       []byte(nil),
			hasErr:     true,
		},
		{
			given:      bufio.NewReader(bytes.NewBuffer([]byte("1,2\n"))),
			recipients: []uint64{1,2},
			body:       []byte(nil),
			hasErr:     true,
		},
	}

	for _, tc := range tt {
		msg := Send{}
		err := msg.Unmarshal(tc.given)
		if tc.hasErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.EqualValues(t, tc.recipients, msg.Recipients)
		assert.Equal(t, tc.body, msg.Body)
	}
}

func TestIncoming_Marshal(t *testing.T) {
	tt := []struct {
		sender uint64
		body   []byte
		want   []byte
	}{
		{
			sender: 1,
			body:   []byte("Hello"),
			want:   []byte("INCOMING\n1\nHello\n"),
		},
		{
			sender: 1,
			body:   []byte(""),
			want:   []byte("INCOMING\n1\n\n"),
		},
		{
			sender: 122,
			body:   []byte("FOOOBAAR\n"),
			want:   []byte("INCOMING\n122\nFOOOBAAR\n\n"),
		},
	}

	for _, tc := range tt {
		msg := NewIncoming(tc.sender, tc.body)
		actual := msg.Marshal()
		assert.Equal(t, tc.want, actual)
	}
}

func TestJoinRecipients(t *testing.T) {
	tt := []struct {
		given []uint64
		want  string
	}{
		{
			[]uint64{1},
			"1",
		},
		{
			[]uint64{},
			"",
		},
		{
			[]uint64{1, 2, 3, 4},
			"1,2,3,4",
		},
	}

	for _, tc := range tt {
		actual := joinRecipients(tc.given)
		assert.Equal(t, tc.want, actual)
	}
}
