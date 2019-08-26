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
