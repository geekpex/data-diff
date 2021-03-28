package main

import (
	"fmt"
	"os"
)

// createSignature creates signature file witch contains chunks of oldFile (a.k.a Basis file)
func createSignature(oldFile *os.File) ([]byte, error) {
	data, err := readFile(oldFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %s", ArgOldFile, err.Error())
	}

	chunks := resolveChunks(data)

	signature, err := writeSignature(chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to write %s file: %s", ArgSignature, err.Error())
	}

	return signature, nil
}
