package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	chunkSeparator = 0x007f
	chunkMinSize   = 31
	chunkMaxSize   = 1023
)

var chunkHash = sha1.New()

type chunk struct {
	start        uint32
	size         uint32
	stopChecksum uint64

	hash []byte

	// For diff processing
	candidates []*chunk
	number     int
}

func NewChunk(data []byte, hash uint64, i, prevIndex int) chunk {
	chunkHash.Reset()
	chunkHash.Write(data[prevIndex : i+1])
	chunkH := chunkHash.Sum(nil)

	if Verbose {
		fmt.Println(base64.StdEncoding.EncodeToString(chunkH))
		fmt.Println(hash, (i - prevIndex + 1), ":", "\""+string(data[prevIndex:i+1])+"\"")
	}

	return chunk{
		start:        uint32(prevIndex),
		size:         uint32(i - prevIndex + 1),
		stopChecksum: hash,
		hash:         chunkH,
	}
}

// writeSignature writes slice of chunks to signature file
func writeSignature(chunks []chunk) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Grow(len(chunks) * (12 + sha1.Size))
	w := io.Writer(buf)

	binary.Write(w, binary.BigEndian, uint32(len(chunks)))
	for i := 0; i < len(chunks); i++ {
		binary.Write(w, binary.BigEndian, chunks[i].start)
		binary.Write(w, binary.BigEndian, chunks[i].size)
		binary.Write(w, binary.BigEndian, chunks[i].stopChecksum)
		w.Write(chunks[i].hash)
	}

	return buf.Bytes(), nil
}

// readSignature reads from r io.Reader slice of chunks that makes a signature file
func readSignature(r io.Reader) (chunks []chunk, err error) {
	var uInt uint32

	err = binary.Read(r, binary.BigEndian, &uInt)
	if err != nil {
		err = fmt.Errorf("failed to read total number of chunks: %s", err.Error())
		return
	}

	chunks = make([]chunk, uInt)
	for i := 0; i < len(chunks); i++ {
		err = binary.Read(r, binary.BigEndian, &chunks[i].start)
		if err != nil {
			err = fmt.Errorf("failed to read [%d] chunk start: %s", i, err.Error())
			return
		}
		err = binary.Read(r, binary.BigEndian, &chunks[i].size)
		if err != nil {
			err = fmt.Errorf("failed to read [%d] chunk size: %s", i, err.Error())
			return
		}
		err = binary.Read(r, binary.BigEndian, &chunks[i].stopChecksum)
		if err != nil {
			err = fmt.Errorf("failed to read [%d] chunk stopChecksum: %s", i, err.Error())
			return
		}
		chunks[i].hash = make([]byte, sha1.Size)
		_, err = r.Read(chunks[i].hash)
		if err != nil {
			err = fmt.Errorf("failed to read [%d] chunk hash: %s", i, err.Error())
			return
		}

		chunks[i].number = i
	}

	return
}

var calcRollingHashFunc = calcRollingHash

func resolveChunks(data []byte) []chunk {
	var hashC = make(chan SingleHash)

	go calcRollingHashFunc(data, hashC)

	var chunks []chunk
	var i, prevIndex int
	var hash uint64
	var sh SingleHash
	for sh = range hashC {
		i, hash = sh.i, sh.h
		if (hash|chunkSeparator) == hash && (i-prevIndex) >= chunkMinSize || (i-prevIndex) == chunkMaxSize {
			// Hash passes chunk separator criterias so mark new chunk
			chunks = append(chunks, NewChunk(data, hash, i, prevIndex))

			if Verbose {
				fmt.Println()
			}

			prevIndex = i + 1
		}
	}

	i = len(data)

	if prevIndex < i {
		// Write last chunk if the last hash was not naturally a chunk separator
		chunks = append(chunks, NewChunk(data, hash, i-1, prevIndex))

		if Verbose {
			fmt.Println()
		}
	}

	return chunks
}
