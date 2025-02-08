package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "size 2",
			size: 2,
		}, {
			name: "size 10",
			size: 10,
		}, {
			name: "size 100",
			size: 100,
		}, {
			name: "size 0",
			size: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str1 := NewRandomString(tt.size)
			str2 := NewRandomString(tt.size)

			assert.Len(t, str1, tt.size)
			assert.Len(t, str2, tt.size)

			if str1 != "" && str2 != "" {
				assert.NotEqual(t, str1, str2)
			} else {
				assert.Empty(t, str1, str2)
			}
		})
	}
}
