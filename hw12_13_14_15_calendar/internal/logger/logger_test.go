package logger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		funcName    string
		message     string
		expectedMsg string
	}{
		{
			name:        "error msg",
			level:       "ERROR",
			funcName:    "Error",
			message:     "This is error message",
			expectedMsg: "ERROR:This is error message",
		},
		{
			name:        "skipp_warn",
			level:       "ERROR",
			funcName:    "Warn",
			message:     "This is error message",
			expectedMsg: "",
		},
		{
			name:        "skipp_info",
			level:       "ERROR",
			funcName:    "Info",
			message:     "This is error message",
			expectedMsg: "",
		},
		{
			name:        "skipp_debug",
			level:       "ERROR",
			funcName:    "Debug",
			message:     "This is error message",
			expectedMsg: "",
		},
		{
			name:        "debug",
			level:       "DEBUG",
			funcName:    "Debug",
			message:     "This is error message",
			expectedMsg: "DEBUG:This is error message",
		},
		{
			name:        "skipp_debug2",
			level:       "INFO",
			funcName:    "Debug",
			message:     "This is error message",
			expectedMsg: "",
		},
	}

	t.Parallel()
	for _, tc := range tests {
		t.Run(fmt.Sprintf("case %s", tc.name), func(t *testing.T) {
			var b bytes.Buffer
			tc := tc
			l := NewLogger(tc.level, &b)

			switch tc.funcName {
			case "Error":
				l.Errorf(tc.message)
			case "Warn":
				l.Warningf(tc.message)
			case "Info":
				l.Infof(tc.message)
			case "Debug":
				l.Debugf(tc.message)
			}

			require.Equal(t, tc.expectedMsg, b.String(), "error output message")
		})
	}
}

func TestFatalf(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		var b bytes.Buffer
		_ = NewLogger("WRONG DB TYPE", &b)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalf")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	var e *exec.ExitError
	require.True(t, err != nil && errors.As(err, &e), "process ran with err %v, want exit status 1", err)
}
