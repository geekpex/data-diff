package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createData(l int, b byte) []byte {
	var d = make([]byte, l)

	for i := 0; i < len(d); i++ {
		d[i] = b
	}

	return d
}

func TestResolveChunks(t *testing.T) {
	var tests = []struct {
		name           string
		data           []byte
		rollingFunc    func(data []byte, out chan<- SingleHash)
		expectedChunks []chunk
	}{
		{
			name: "0x007f hashes. Min size chunks",
			data: createData(256, 0x00),
			rollingFunc: func(data []byte, out chan<- SingleHash) {
				for i := windowSize - 1; i < len(data); i++ {
					out <- SingleHash{i: i, h: 0x007f}
				}
				close(out)
			},
			expectedChunks: []chunk{
				{
					start:        0,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        32,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        64,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        96,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        128,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        160,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        192,
					size:         32,
					stopChecksum: 0x007f,
				},
				{
					start:        224,
					size:         32,
					stopChecksum: 0x007f,
				},
			},
		},
		{
			name: "0x00 hashes. Max size chunks",
			data: createData(1200, 0x00),
			rollingFunc: func(data []byte, out chan<- SingleHash) {
				for i := windowSize - 1; i < len(data); i++ {
					out <- SingleHash{i: i, h: 0x00}
				}
				close(out)
			},
			expectedChunks: []chunk{
				{
					start:        0,
					size:         1024,
					stopChecksum: 0x00,
				},
				{
					start:        1024,
					size:         176,
					stopChecksum: 0x00,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calcRollingHashFunc = tt.rollingFunc

			gotChunks := resolveChunks(tt.data)

			if len(gotChunks) != len(tt.expectedChunks) {
				assert.FailNowf(t, "Expected amount of chunks should be equal to received ones", "%d != %d", len(gotChunks), len(tt.expectedChunks))
			}

			for i := 0; i < len(tt.expectedChunks); i++ {
				// We are not interested in hash field's value because it is calculated with sha1 hashing alg and that is
				// tested elsewhere and trusted to work.
				assert.Equal(t, tt.expectedChunks[i].size, gotChunks[i].size, "Chunk sizes should be equal")
				assert.Equal(t, tt.expectedChunks[i].start, gotChunks[i].start, "Chunk start should be equal")
				assert.Equal(t, tt.expectedChunks[i].stopChecksum, gotChunks[i].stopChecksum, "Chunk stopChecksum should be equal")
			}
		})
	}

	calcRollingHashFunc = calcRollingHash
}
