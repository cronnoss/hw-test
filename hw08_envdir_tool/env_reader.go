package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	environment := make(Environment)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := file.Name()

		if strings.Contains(fileName, "=") {
			continue
		}

		filePath := dir + "/" + fileName

		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}
		if stat.Size() == 0 {
			environment[fileName] = EnvValue{"", true}
			continue
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		scanner.Scan()
		str := scanner.Text()
		str = strings.TrimRight(str, " \t")
		str = string(bytes.ReplaceAll([]byte(str), []byte("\x00"), []byte("\n")))

		environment[fileName] = EnvValue{str, false}

		err = f.Close()
		if err != nil {
			return nil, err
		}
	}

	return environment, nil
}
