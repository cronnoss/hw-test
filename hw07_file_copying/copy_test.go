package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("Write file completely as is", func(t *testing.T) {
		inputFile := "inputFile.txt"
		outFile := "outFile.txt"
		testData := []byte("Hello, world!")

		prepareTestFiles(t, inputFile, testData)

		err := Copy(inputFile, outFile, 0, 0)
		require.NoError(t, err)

		assertData(t, outFile, testData)

		removeTestFiles(t, inputFile, outFile)
	})

	t.Run("Write file with offset 7 limit 0", func(t *testing.T) {
		inputFile := "inputFile1.txt"
		outFile := "outFile1.txt"
		testData := []byte("Hello, world!")

		prepareTestFiles(t, inputFile, testData)

		err := Copy(inputFile, outFile, 7, 0)
		require.NoError(t, err)

		assertData(t, outFile, testData[7:])

		removeTestFiles(t, inputFile, outFile)
	})

	t.Run("Write file with offset 0 limit 7", func(t *testing.T) {
		inputFile := "inputFile2.txt"
		outFile := "outFile2.txt"
		testData := []byte("Hello, world!")

		prepareTestFiles(t, inputFile, testData)

		err := Copy(inputFile, outFile, 0, 7)
		require.NoError(t, err)

		assertData(t, outFile, testData[:7])

		removeTestFiles(t, inputFile, outFile)
	})

	t.Run("Write file with offset 7 limit 5", func(t *testing.T) {
		inputFile := "inputFile3.txt"
		outFile := "outFile3.txt"
		testData := []byte("Hello, world!")

		prepareTestFiles(t, inputFile, testData)

		err := Copy(inputFile, outFile, 7, 5)
		require.NoError(t, err)

		assertData(t, outFile, testData[7:12])

		removeTestFiles(t, inputFile, outFile)
	})

	t.Run("Write file with offset exceeds file size", func(t *testing.T) {
		inputFile := "inputFile4.txt"
		outFile := "outFile4.txt"
		testData := []byte("Hello, world!")

		prepareTestFiles(t, inputFile, testData)

		err := Copy(inputFile, outFile, 100, 0)
		require.Equal(t, ErrOffsetExceedsFileSize, err)

		removeTestFiles(t, inputFile)
	})

	t.Run("Unsupported file", func(t *testing.T) {
		inputFile := "inputFile5.txt"
		outFile := "outFile5.txt"

		prepareTestFiles(t, inputFile, []byte{})

		err := Copy(inputFile, outFile, 0, 0)
		require.Truef(t, errors.Is(err, ErrUnsupportedFile), "actual err - %v", err)

		removeTestFiles(t, inputFile)
	})

	t.Run("File does not exist", func(t *testing.T) {
		inputFile := "inputFile6.txt"
		outFile := "outFile6.txt"

		err := Copy(inputFile, outFile, 0, 0)
		require.Equal(t, "file does not exist", err.Error())
	})
}

func prepareTestFiles(t *testing.T, inputFile string, testData []byte) {
	t.Helper()
	err := os.WriteFile(inputFile, testData, os.ModePerm)
	require.NoError(t, err)
}

func removeTestFiles(t *testing.T, files ...string) {
	t.Helper()
	for _, file := range files {
		err := os.Remove(file)
		require.NoError(t, err)
	}
}

func assertData(t *testing.T, outFile string, expectedData []byte) {
	t.Helper()
	outData, err := os.ReadFile(outFile)
	require.NoError(t, err)
	require.Equal(t, expectedData, outData)
}
