package main

import (
	"bytes"
	"fmt"
	"io"
)

// DeltaBuffer represents buffer that contains the delta of basis and changed file
type DeltaBuffer interface {
	// Bytes closes the buffer and returns bytes from buffer
	Bytes() []byte

	// AddLiteral writes literal command to buffer
	AddLiteral(data []byte)

	// AddCopy writes copy command to buffer
	AddCopy(start, length uint64)
}

// declared in global level for unit tests
var deltaBufferConstructor = NewRdiffDelta

// createDelta processed signature and newfile to create delta which contains changes between new file and basis file
// from which the signature was created
func createDelta(signature, newFile io.Reader) ([]byte, error) {
	chunks, err := readSignature(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %s", ArgSignature, err.Error())
	}

	data, err := readFile(newFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %s", ArgNewFile, err.Error())
	}

	var deltaB = deltaBufferConstructor()

	newChunks := resolveChunks(data)

	if Verbose {
		fmt.Println()
		fmt.Println("Finding differences:")
		fmt.Println()
	}

	for i := 0; i < len(newChunks); i++ {
		for j := 0; j < len(chunks); j++ {
			if chunks[j].stopChecksum == newChunks[i].stopChecksum {
				newChunks[i].candidates = append(newChunks[i].candidates, &chunks[j])
			}
		}
	}

	for i := 0; i < len(newChunks); i++ {
		eq := false
		for j := 0; j < len(newChunks[i].candidates); j++ {
			c := newChunks[i].candidates[j]

			eq = bytes.Equal(newChunks[i].hash, c.hash)

			if eq {
				deltaB.AddCopy(uint64(c.start), uint64(c.size))

				if Verbose {
					fmt.Println(i, "matches chunk in basefile:", c.number)
				}
				break
			}
		}

		if !eq {
			deltaB.AddLiteral(data[newChunks[i].start : newChunks[i].start+newChunks[i].size])
			if Verbose {
				fmt.Println(i, "No matching chunk in basefile. Content:", "\""+string(data[newChunks[i].start:newChunks[i].start+newChunks[i].size])+"\"")
			}
		}
	}

	return deltaB.Bytes(), nil
}
