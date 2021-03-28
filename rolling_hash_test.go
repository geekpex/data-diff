package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcRollingHash(t *testing.T) {
	var tests = []struct {
		name       string
		data       []byte
		assertHash func(t *testing.T, n uint64)
	}{
		{
			name: "256 slice with only 0x00 values",
			data: (func() []byte {
				var d = make([]byte, 256)

				for i := 0; i < len(d); i++ {
					d[i] = 0x00
				}

				return d
			}()),
			assertHash: func(t *testing.T, n uint64) {
				assert.Equal(t, uint64(0), n, "hash should be equal to 0")
			},
		},
		{
			name: "256 slice with only 0xff values",
			data: (func() []byte {
				var d = make([]byte, 256)

				for i := 0; i < len(d); i++ {
					d[i] = 0xff
				}

				return d
			}()),
			assertHash: func(t *testing.T, n uint64) {
				assert.Equal(t, uint64(0x21eabe90), n, "hash should be equal to 0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan SingleHash)

			go calcRollingHash(tt.data, ch)

			for n := range ch {
				tt.assertHash(t, n.h)
			}
		})
	}
}
