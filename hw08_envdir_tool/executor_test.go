package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	cases := []struct {
		name     string
		cmd      []string
		expected int
	}{
		{
			name:     "1",
			cmd:      []string{"/bin/bash", "./testdata/echo.sh", "arg1=1", "arg2=2"},
			expected: 0,
		},
		{
			name:     "1",
			cmd:      []string{"/bin/bash", "./testdata/echo.s", "arg1=1", "arg2=2"},
			expected: 127,
		},
		{
			name: "1",
			cmd: []string{
				"test",
			},
			expected: 1,
		},
		{
			name: "0",
			cmd: []string{
				"test",
				"-test",
			},
			expected: 0,
		},
		{
			name: "-1",
			cmd: []string{
				"-test",
			},
			expected: -1,
		},
	}

	env, err := ReadDir("./testdata/env")
	require.Nil(t, err)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ComArgVar := RunCmd(c.cmd, env)
			require.Equal(t, c.expected, ComArgVar)
		})
	}
}
