package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

const bufferSize = 1024

func Copy(fromPath, toPath string, offset, limit int64) error {
	inFile, err := os.Open(fromPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file does not exist")
		}
		return ErrUnsupportedFile
	}

	fileInfo, err := inFile.Stat()
	if err != nil {
		return ErrUnsupportedFile
	}

	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return ErrUnsupportedFile
	}
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	outFile, err := os.Create(toPath)
	if err != nil {
		return err
	}

	if limit == 0 {
		limit = fileSize
	}
	if (fileSize - offset) < limit {
		limit = fileSize - offset
	}

	bar := pb.StartNew(int(limit))
	buf := make([]byte, bufferSize)

	for limit > 0 {
		readBytes, err := inFile.ReadAt(buf, offset)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		writeBytes := readBytes
		if int64(writeBytes) > limit {
			writeBytes = int(limit)
		}
		_, err = outFile.Write(buf[:writeBytes])
		if err != nil {
			return err
		}

		limit -= int64(readBytes)
		offset += int64(bufferSize)
		bar.Add(writeBytes)
	}

	bar.Finish()
	return outFile.Chmod(fileInfo.Mode().Perm())
}
