package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	dir := "./testdata/env"
	expected := Environment{
		"BAR":   EnvValue{"bar", false},
		"EMPTY": EnvValue{"", false},
		"FOO":   EnvValue{"   foo\nwith new line", false},
		"HELLO": EnvValue{"\"hello\"", false},
		"UNSET": EnvValue{"", true},
	}

	t.Run("ReadDir", func(t *testing.T) {
		envs, err := ReadDir(dir)

		require.Nil(t, err)
		require.Equal(t, expected, envs)
	})

	t.Run("DirNotExist", func(t *testing.T) {
		_, err := ReadDir("testdata/env1")

		require.NotNil(t, err)
	})

	t.Run("File name contains '='", func(t *testing.T) {
		fileName, err := os.CreateTemp(dir, "FOO=")
		if err != nil {
			fmt.Println(err)
			return
		}

		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Println(err)
			}
		}(fileName.Name())

		envs, err := ReadDir(dir)

		require.Nil(t, err)
		require.Equal(t, expected, envs)
	})
}
