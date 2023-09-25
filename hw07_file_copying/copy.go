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

type barWriter struct {
	w io.Writer
	b *pb.ProgressBar
}

func (bw *barWriter) Write(p []byte) (n int, err error) {
	n, err = bw.w.Write(p)
	bw.b.Add(n)
	return
}

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

	_, err = inFile.Seek(offset, 0)
	if err != nil {
		return ErrUnsupportedFile
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

	bar := pb.Full.Start64(limit)
	defer bar.Finish()

	_, err = io.CopyN(&barWriter{outFile, bar}, inFile, limit)
	if err != nil {
		return err
	}

	return outFile.Chmod(fileInfo.Mode().Perm())
}
