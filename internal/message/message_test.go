package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
			[]uint64{1,2,3,4},
			"1,2,3,4",
		},
	}

	for _, tc := range tt {
		actual := joinRecipients(tc.given)
		assert.Equal(t, tc.want, actual)
	}
}
