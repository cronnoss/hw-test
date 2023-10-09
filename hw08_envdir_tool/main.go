package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Too few arguments, expected the format: go-envdir /path/to/env/dir command arg1 arg2")
		os.Exit(1)
	}

	environment, err := ReadDir(os.Args[1])
	if err != nil {
		fmt.Println("Failed to read directory: ", err)
		os.Exit(1)
	}

	returnCode := RunCmd(os.Args[2:], environment)
	os.Exit(returnCode)
}
