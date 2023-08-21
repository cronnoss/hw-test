package hw02unpackstring

import (
	"errors"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	runeArray := []rune(s)

	newRuneArray := []rune{}
	if (len(runeArray) > 0) && (unicode.IsDigit(runeArray[0])) {
		return "", ErrInvalidString
	}

	for i := 0; i < len(runeArray); i++ {
		tempRune := runeArray[i]
		if (i+1 < len(runeArray)) && (unicode.IsDigit(runeArray[i+1])) {
			nextRuneIsDigit := i + 1
			if (i+2 < len(runeArray)) && (unicode.IsDigit(runeArray[i+2])) {
				return "", ErrInvalidString
			}
			differenceBetweenDigitRuneAndZero := int(runeArray[nextRuneIsDigit] - '0')
			for j := 0; j < differenceBetweenDigitRuneAndZero; j++ {
				newRuneArray = append(newRuneArray, tempRune)
			}
			i++
		} else {
			newRuneArray = append(newRuneArray, tempRune)
		}
	}

	return string(newRuneArray), nil
}
