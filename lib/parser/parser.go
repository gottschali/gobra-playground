package parser

import (
	"regexp"
	"strconv"
)

type Position struct {
	Line int `json:"line"`
	Char int `json:"char"`
}

type VerificationError struct {
	Message  string   `json:"message"`
	Position Position `json:"Position"`
}

// where stats
type VerificationResponse struct {
	Verified bool                `json:"verified"`
	Timeout  bool                `json:"timeout"`
	Errors   []VerificationError `json:"errors"`
	// output: list[error string] `json:"output"`
	Duration float64 `json:"duration"`
	Stats    string  `json:"stats"`
}

var error_re = regexp.MustCompile(`- Error at: <(.*):([0-9]+):([0-9]+)> (.*)\n(.*)`)
var num_error_re = regexp.MustCompile(`\[main\] INFO viper.gobra.Gobra - Gobra has found ([0-9]+) error`)
var main_error_re = regexp.MustCompile(`^[^0-9][^0-9].*`)

func ParseGobraOutput(output string) (VerificationResponse, error) {
	r := VerificationResponse{}
	r.Errors = make([]VerificationError, 0)
	for _, match := range error_re.FindAllStringSubmatch(output, -1) {
		line, err := strconv.Atoi(match[2])
		if err != nil {
			return r, err
		}
		char, err := strconv.Atoi(match[3])
		if err != nil {
			return r, err
		}
		message := match[4] + "\n" + match[5]
		r.Errors = append(r.Errors, VerificationError{
			Message:  message,
			Position: Position{line, char},
		})
	}
	r.Verified = len(r.Errors) == 0
	return r, nil
}
