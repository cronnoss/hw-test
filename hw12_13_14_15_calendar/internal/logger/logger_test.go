package logger

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name           string
		level          string
		funcToCall     func(*Logger)
		expectedOutput string
	}{
		{
			name:           "Error Level - Errorf",
			level:          "ERROR",
			funcToCall:     func(l *Logger) { l.Errorf("Test Error") },
			expectedOutput: "ERROR:Test Error",
		},
		{
			name:           "Error Level - Warningf",
			level:          "ERROR",
			funcToCall:     func(l *Logger) { l.Warningf("Test Warning") },
			expectedOutput: "",
		},
		{
			name:           "Error Level - Infof",
			level:          "ERROR",
			funcToCall:     func(l *Logger) { l.Infof("Test Info") },
			expectedOutput: "",
		},
		{
			name:           "Error Level - Debugf",
			level:          "ERROR",
			funcToCall:     func(l *Logger) { l.Debugf("Test Debug") },
			expectedOutput: "",
		},
		{
			name:           "Warn Level - Errorf",
			level:          "WARN",
			funcToCall:     func(l *Logger) { l.Errorf("Test Error") },
			expectedOutput: "ERROR:Test Error",
		},
		{
			name:           "Warn Level - Warningf",
			level:          "WARN",
			funcToCall:     func(l *Logger) { l.Warningf("Test Warning") },
			expectedOutput: "WARN:Test Warning",
		},
		{
			name:           "Debug Level - Infof",
			level:          "DEBUG",
			funcToCall:     func(l *Logger) { l.Infof("Test Info") },
			expectedOutput: "INFO:Test Info",
		},
		{
			name:           "Info Level - Debugf",
			level:          "INFO",
			funcToCall:     func(l *Logger) { l.Debugf("Test Debug") },
			expectedOutput: "",
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger, _ := New(tt.level, &output)
			tt.funcToCall(logger)
			if output.String() != tt.expectedOutput {
				t.Errorf("got %q, want %q", output.String(), tt.expectedOutput)
			}
		})
	}

	t.Run("Wrong level", func(t *testing.T) {
		var output bytes.Buffer
		_, err := New("WRONG", &output)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestLogger_Fatalf(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		var output bytes.Buffer
		logger, _ := New("DEBUG", &output)
		logger.Fatalf("Test Fatal")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatalf")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	var e *exec.ExitError
	if errors.As(err, &e) && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
