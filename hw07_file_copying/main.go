package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()

	if err := validateInput(from, to, limit, offset); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := Copy(from, to, offset, limit); err != nil {
		fmt.Println(err)
	}
}

func validateInput(from, to string, limit, offset int64) error {
	if from == "" {
		return errors.New("from flag is required")
	}
	if to == "" {
		return errors.New("to flag is required")
	}
	if limit < 0 {
		return errors.New("limit flag must be greater than or equal to zero")
	}
	if offset < 0 {
		return errors.New("offset flag must be greater than or equal to zero")
	}
	return nil
}
