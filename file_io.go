package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// writeFile writes to file pointed by argFile2 global variable
func writeFile(data []byte) error {
	return ioutil.WriteFile(argOutputFile, data, os.ModePerm)
}

// openReadFile opens file poinsted by name.
// If file does not exist or open fails error is returned.
func openReadFile(arg, name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s file does not exist: %s", arg, name)
		}
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	if stat.IsDir() {
		file.Close()
		return nil, fmt.Errorf("%s is a directory", arg)
	}

	return file, nil
}

// checkFileDoesNotExist checks that file pointed by name does not already exist.
func checkFileDoesNotExist(arg, name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	file.Close()
	return fmt.Errorf("%s file file already exists: %s", arg, name)
}

// readFile reads a files content and returns error if it is a directory
func readFile(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
