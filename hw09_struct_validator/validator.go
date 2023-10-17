package hw09structvalidator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	validateTag = "validate"
	lenRule     = "len"
	regexpRule  = "regexp"
	inRule      = "in"
	minRule     = "min"
	maxRule     = "max"
)

type IllegalArgumentError struct {
	Message string
}

func (i IllegalArgumentError) Error() string {
	return fmt.Sprintf("illegal argument: %s", i.Message)
}

func NewIllegalArgumentError(message string) IllegalArgumentError {
	return IllegalArgumentError{
		Message: message,
	}
}

type FieldValidationError struct {
	Message string
}

func (f FieldValidationError) Error() string {
	return f.Message
}

func NewFieldValidationError(message string) FieldValidationError {
	return FieldValidationError{
		Message: message,
	}
}

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errStr strings.Builder
	errStr.Write([]byte("Validation Errors:\n"))
	for _, e := range v {
		errStr.Write([]byte(fmt.Sprintf("Field: %s, Err: %s\n", e.Field, e.Err)))
	}
	return errStr.String()
}

func NewValidationErrors(validationErrors ...ValidationError) ValidationErrors {
	return validationErrors
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return NewIllegalArgumentError("expected a struct")
	}

	var valErrs ValidationErrors
	for _, f := range reflect.VisibleFields(val.Type()) {
		if err := validate(f, val, &valErrs); err != nil {
			return err
		}
	}
	if len(valErrs) > 0 {
		return valErrs
	}
	return nil
}

func validate(f reflect.StructField, r reflect.Value, errs *ValidationErrors) error {
	validationTagValue := f.Tag.Get(validateTag)
	if validationTagValue == "" {
		return nil
	}

	field := r.FieldByName(f.Name)

	rules, err := makeRules(validationTagValue)
	if err != nil {
		return err
	}

	if field.Kind() == reflect.String {
		if err := validateString(rules, f.Name, field.String(), errs); err != nil {
			return err
		}
	}
	if field.Kind() == reflect.Int {
		if err := validateInt(rules, f.Name, field.Interface().(int), errs); err != nil {
			return err
		}
	}
	if field.Kind() == reflect.Slice {
		if err := validateSlice(f, field, rules, errs); err != nil {
			return err
		}
	}
	return nil
}

func validateSlice(f reflect.StructField, field reflect.Value, rules map[string]string, errs *ValidationErrors) error {
	if field.Type().Elem().Kind() == reflect.String {
		if err := validateStringArr(field, rules, f, errs); err != nil {
			return err
		}
	}
	if field.Type().Elem().Kind() == reflect.Int {
		if err := validateIntArr(field, rules, f, errs); err != nil {
			return err
		}
	}
	return nil
}

func validateIntArr(v reflect.Value, rules map[string]string, f reflect.StructField, errs *ValidationErrors) error {
	elems := v.Interface().([]int)
	for _, el := range elems {
		err := validateInt(rules, f.Name, el, errs)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateStringArr(v reflect.Value, rules map[string]string, f reflect.StructField, errs *ValidationErrors) error {
	elems := v.Interface().([]string)
	for _, el := range elems {
		err := validateString(rules, f.Name, el, errs)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateString(rules map[string]string, fieldName, fieldValue string, errs *ValidationErrors) error {
	for ruleName, ruleArg := range rules {
		switch ruleName {
		case lenRule:
			l, err := getIntValue(fieldName, ruleArg)
			if err != nil {
				return err
			}
			fLen := len(fieldValue)
			if fLen != l {
				*errs = append(*errs, ValidationError{
					Field: fieldName,
					Err:   NewFieldValidationError(fmt.Sprintf("len must be %d, got %d for '%s'", l, fLen, fieldValue)),
				})
			}
		case regexpRule:
			reg, err := regexp.Compile(ruleArg)
			if err != nil {
				return NewIllegalArgumentError(
					fmt.Sprintf("illegal regexp value for filed '%s', must be valid regexp, got '%s'", fieldName, ruleArg),
				)
			}
			if !reg.Match([]byte(fieldValue)) {
				*errs = append(*errs, ValidationError{
					Field: fieldName,
					Err: NewFieldValidationError(
						fmt.Sprintf("fieldValue must match regexp '%s', actual value '%s'", ruleArg, fieldValue),
					),
				})
			}
		case inRule:
			variants := strings.Split(ruleArg, ",")
			if !slices.Contains(variants, fieldValue) {
				*errs = append(*errs, ValidationError{
					Field: fieldName,
					Err: NewFieldValidationError(
						fmt.Sprintf("fieldValue must be one of %s values, given '%s'", variants, fieldValue),
					),
				})
			}
		default:
			return NewIllegalArgumentError(
				fmt.Sprintf("unsupported validation '%s' for string filed '%s'", ruleName, fieldName),
			)
		}
	}
	return nil
}

func validateInt(rules map[string]string, fieldName string, fieldValue int, errs *ValidationErrors) error {
	for ruleName, ruleArg := range rules {
		switch ruleName {
		case minRule:
			minimum, err := getIntValue(fieldName, ruleArg)
			if err != nil {
				return err
			}
			if fieldValue < minimum {
				*errs = append(*errs, ValidationError{
					Field: fieldName,
					Err:   NewFieldValidationError(fmt.Sprintf("minimum value %d, got %d", minimum, fieldValue)),
				})
			}
		case maxRule:
			maximum, err := getIntValue(fieldName, ruleArg)
			if err != nil {
				return err
			}
			if fieldValue > maximum {
				*errs = append(*errs, ValidationError{
					Field: fieldName,
					Err:   NewFieldValidationError(fmt.Sprintf("maximum value %d, got %d", maximum, fieldValue)),
				})
			}
		case inRule:
			var variants []int
			for _, v := range strings.Split(ruleArg, ",") {
				variant, err := getIntValue(fieldName, v)
				if err != nil {
					return err
				}
				variants = append(variants, variant)
			}
			if !slices.Contains(variants, fieldValue) {
				*errs = append(*errs, ValidationError{
					Field: fieldName,
					Err: NewFieldValidationError(
						fmt.Sprintf("fieldValue must be one of %v values, given %d", variants, fieldValue),
					),
				})
			}
		default:
			return NewIllegalArgumentError(
				fmt.Sprintf("unsupported validation '%s' for string filed '%s'", ruleName, fieldName),
			)
		}
	}
	return nil
}

func getIntValue(fieldName, ruleArg string) (int, error) {
	l, err := strconv.Atoi(ruleArg)
	if err != nil {
		return l, NewIllegalArgumentError(
			fmt.Sprintf("illegal value in tag for filed '%s', must be number, got '%s'", fieldName, ruleArg),
		)
	}
	return l, nil
}

func makeRules(validationTag string) (map[string]string, error) {
	rules := make(map[string]string)
	for _, rule := range strings.Split(validationTag, "|") {
		separated := strings.Split(rule, ":")
		if len(separated) != 2 {
			return nil, NewIllegalArgumentError("validation tag must be in form key:value")
		}
		rules[separated[0]] = separated[1]
	}
	return rules, nil
}
