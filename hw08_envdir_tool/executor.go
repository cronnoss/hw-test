package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		fmt.Println("Too few arguments, expected the format: go-envdir /path/to/env/dir command arg1 arg2")
		return 1
	}

	command := exec.Command(cmd[0], cmd[1:]...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	for key, value := range env {
		if value.NeedRemove {
			err := os.Unsetenv(key)
			if err != nil {
				fmt.Println("Failed to unset environment variable: ", err.Error())
				return 1
			}
		} else {
			err := os.Setenv(key, value.Value)
			if err != nil {
				fmt.Println("Failed to set environment variable: ", err.Error())
				return 1
			}
		}
	}

	err := command.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
	}

	return command.ProcessState.ExitCode()
}
