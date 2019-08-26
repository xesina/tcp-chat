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
