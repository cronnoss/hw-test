package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errStr string
	for _, e := range v {
		errStr += fmt.Sprintf("Field: %s Error: %v\n", e.Field, e.Err)
	}
	return errStr
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Struct {
		return errors.New("expected a struct")
	}
	var valErrs ValidationErrors
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		tag := field.Tag.Get("validate")
		if tag != "" && fieldValue.CanInterface() {
			err := validateField(tag, fieldValue.Interface())
			if err != nil {
				valErrs = append(valErrs, ValidationError{Field: field.Name, Err: err})
			}
		}
	}
	if len(valErrs) > 0 {
		return valErrs
	}
	return nil
}

func extractRule(rule string) (string, string, error) {
	split := strings.Split(rule, ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid rule format: %s", rule)
	}
	return split[0], split[1], nil
}

func validateLen(value interface{}, ruleValue string) error {
	expectedLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return err
	}
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		return validateLenSlice(value, expectedLen)
	}
	return validateLenString(value, expectedLen)
}

func validateLenSlice(value interface{}, expectedLen int) error {
	s := reflect.ValueOf(value)
	for i := 0; i < s.Len(); i++ {
		if len(s.Index(i).String()) != expectedLen {
			return fmt.Errorf("expected length of %d, got %d", expectedLen, len(s.Index(i).String()))
		}
	}
	return nil
}

func validateLenString(value interface{}, expectedLen int) error {
	if len(value.(string)) != expectedLen {
		return fmt.Errorf("expected length of %d, got %d", expectedLen, len(value.(string)))
	}
	return nil
}

func validateRegexp(value interface{}, ruleValue string) error {
	s := reflect.ValueOf(value).String()
	matched, err := regexp.Compile(ruleValue)
	if err != nil {
		return err
	}
	if !matched.MatchString(s) {
		return fmt.Errorf("string does not match regexp %s", ruleValue)
	}
	return nil
}

func validateMin(value interface{}, ruleValue string) error {
	minimum, err := strconv.Atoi(ruleValue)
	if err != nil {
		return err
	}
	if value.(int) < minimum {
		return fmt.Errorf("number is less than minimum of %d", minimum)
	}
	return nil
}

func validateMax(value interface{}, ruleValue string) error {
	maximum, err := strconv.Atoi(ruleValue)
	if err != nil {
		return err
	}
	if value.(int) > maximum {
		return fmt.Errorf("number is more than maximum of %d", maximum)
	}
	return nil
}

func validateIn(value interface{}, ruleValue string) error {
	split := strings.Split(ruleValue, ",")
	switch reflect.TypeOf(value).Kind() { //nolint:exhaustive
	case reflect.String:
		return validateInString(value, split)
	case reflect.Int:
		return validateInInt(value, split)
	default:
		return fmt.Errorf("unknown type for 'in' rule")
	}
}

func validateInString(value interface{}, split []string) error {
	s := reflect.ValueOf(value).String()
	if !stringInSlice(s, split) {
		return fmt.Errorf("string %s is not in set %v", s, split)
	}
	return nil
}

func validateInInt(value interface{}, split []string) error {
	n := reflect.ValueOf(value).Int()
	ns := make([]int, len(split))
	for i, v := range split {
		ns[i], _ = strconv.Atoi(v)
	}
	if !intInSlice(int(n), ns) {
		return fmt.Errorf("number %d is not in set %v", n, ns)
	}
	return nil
}

func validateField(tag string, value interface{}) error {
	rules := strings.Split(tag, "|")
	for _, rule := range rules {
		ruleName, ruleValue, err := extractRule(rule)
		if err != nil {
			return err
		}
		switch ruleName {
		case "len":
			err = validateLen(value, ruleValue)
		case "regexp":
			err = validateRegexp(value, ruleValue)
		case "min":
			err = validateMin(value, ruleValue)
		case "max":
			err = validateMax(value, ruleValue)
		case "in":
			err = validateIn(value, ruleValue)
		default:
			err = fmt.Errorf("unknown tag: %s", ruleName)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func stringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func intInSlice(n int, slice []int) bool {
	for _, v := range slice {
		if v == n {
			return true
		}
	}
	return false
}
