package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

//go:generate easyjson -all stats.go
type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	scanner := bufio.NewScanner(r)
	stat := make(DomainStat)
	var user User
	for scanner.Scan() {
		if err := user.UnmarshalJSON(scanner.Bytes()); err != nil {
			return nil, fmt.Errorf("error unmarshalling user: %w", err)
		}
		if strings.HasSuffix(user.Email, domain) {
			stat[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return stat, nil
}
