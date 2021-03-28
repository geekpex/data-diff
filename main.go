package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	argMode       string
	argOutputFile string

	Verbose       = false
	ForceOverride = false
)

const (
	ModeSignature = "signature"
	ModeDelta     = "delta"

	ArgSignature = "SIGNATURE"
	ArgDelta     = "DELTA"
	ArgNewFile   = "NEWFILE"
	ArgOldFile   = "BASIS"

	usageText = `
Usage: data-diff [OPTIONS] signature [BASIS [SIGNATURE]]
                 [OPTIONS] delta SIGNATURE [NEWFILE [DELTA]]

Options:
-v, --verbose             Trace internal processing
-?, --help                Show this help message
-f, --force               Force overwriting existing files`

	noArgumentsText = "You must specify an action: `signature' or `delta'." +
		"\nTry `data-diff --help' for more information."
)

// processFileArg checks file arguments details
func processFileArg(args []string, idx int, argName string, read bool) (readFile *os.File, err error) {
	if len(args) <= idx {
		return nil, fmt.Errorf("argument \"%s\" is missing", argName)
	}

	if read {
		readFile, err = openReadFile(argName, args[idx])
		if err != nil {
			return nil, err
		}
	} else if !ForceOverride {
		err = checkFileDoesNotExist(argName, args[idx])
		if err != nil {
			return nil, err
		}
	}

	return
}

// processArguments processes passed arguments and returns opened files handlers if it succeeds
// otherwise an error is returned.
func processArguments(args []string) (file0, file1 *os.File, err error) {
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case ModeSignature:
			argMode = ModeSignature
		case ModeDelta:
			argMode = ModeDelta
		default:
			return nil, nil, fmt.Errorf("unsupported mode: %s", os.Args[1])
		}
	} else {
		return nil, nil, fmt.Errorf("first argument is missing")
	}

	defer func() {
		if err != nil && file0 != nil {
			file0.Close()
		}
		if err != nil && file1 != nil {
			file1.Close()
		}
	}()

	switch argMode {
	case ModeSignature:
		file0, err = processFileArg(args, 1, ArgOldFile, true)
		if err != nil {
			return
		}

		_, err = processFileArg(args, 2, ArgSignature, false)
		if err != nil {
			return
		}
		argOutputFile = args[2]
	case ModeDelta:
		file0, err = processFileArg(args, 1, ArgSignature, true)
		if err != nil {
			return
		}

		file1, err = processFileArg(args, 2, ArgNewFile, true)
		if err != nil {
			return
		}

		_, err = processFileArg(args, 3, ArgDelta, false)
		if err != nil {
			return
		}
		argOutputFile = args[3]
	}

	return
}

// stdErr prints params to stderr
func stdErr(params ...interface{}) {
	fmt.Fprintln(os.Stderr, params...)
}

func main() {
	if len(os.Args) == 1 {
		stdErr(noArgumentsText)
		os.Exit(1)
	}

	var args []string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-?", "--help":
			fmt.Println(usageText)
			os.Exit(0)
		case "-v", "--verbose":
			Verbose = true
		case "-f", "--force":
			ForceOverride = true
		default:
			if len(arg) > 0 && arg[0] == '-' {
				stdErr("data-diff: unknown option:", arg)
				os.Exit(2)
			}
			args = append(args, arg)
		}
	}

	file0, file1, err := processArguments(args)
	if err != nil {
		stdErr("data-diff:", err.Error())
		os.Exit(2)
	}

	// Run in specified mode
	var output []byte
	switch argMode {
	case ModeSignature:
		output, err = createSignature(file0)

		file0.Close()
	case ModeDelta:
		output, err = createDelta(file0, file1)

		file0.Close()
		file1.Close()
	}

	if err != nil {
		stdErr("data-diff:", err.Error())
		os.Exit(3)
	}

	err = writeFile(output)
	if err != nil {
		stdErr("data-diff:", err.Error())
		os.Exit(4)
	}

	os.Exit(0)
}
