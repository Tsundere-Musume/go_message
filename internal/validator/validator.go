package validator

import "regexp"

// TODO: use better validaiton
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9\\.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if _, ok := v.FieldErrors[key]; !ok {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return len(value) > 0
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func MinChars(value string, length int) bool {
	return len(value) >= length
}
