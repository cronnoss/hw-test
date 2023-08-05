package hw02unpackstring

import (
	"errors"
)

var ErrInvalidString = errors.New("invalid string")

var isDigit = func(c byte) bool {
	return c >= '0' && c <= '9'
}

func Unpack(s string) (string, error) {
	var result string
	for i := 0; i < len(s); i++ {
		if isDigit(s[i]) {
			err := handleDigit(&result, s, i)
			if err != nil {
				return "", err
			}
		} else {
			result += string(s[i])
		}
	}
	return result, nil
}

func handleDigit(result *string, s string, i int) error {
	if i == 0 {
		return ErrInvalidString
	}
	if s[i] == '0' {
		if s[i-1] == '1' {
			return ErrInvalidString
		}
		*result = (*result)[:len(*result)-1]
	}
	for j := 0; j < int(s[i]-'0')-1; j++ {
		*result += string(s[i-1])
	}
	return nil
}
