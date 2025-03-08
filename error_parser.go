package main

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var num_error_re = regexp.MustCompile(`\[main\] INFO viper.gobra.Gobra - Gobra has found ([0-9]+) error`)
var main_error_re = regexp.MustCompile(`^[^0-9][^0-9].*`)

func ParseGobraOutput(output string) (VerificationResponse, error) {
	r := VerificationResponse{}

	num_error_match := num_error_re.FindStringSubmatch(output)
	if len(num_error_match) < 2 {
		return r, errors.New("Failed to extract the number of errors")
	}
	num_error, err := strconv.Atoi(num_error_match[1])
	if err != nil {
		return r, err
	}
	r.Verified = num_error == 0

	first_error := strings.Index(output, "ERROR")
	var main_error string
	if first_error != -1 {
		next_line := strings.Index(output[first_error:], "\n")
		main_error = main_error_re.FindString(output[first_error+next_line+1:])
	}
	r.MainError = VerificationError{ErrorMessage: main_error}

	return r, nil
}
