package parser

import (
	"errors"
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
	Stats    []any   `json:"stats"`
}

var numErrorRe = regexp.MustCompile(`Gobra has found ([0-9]+) error\(s\)`)
var errorWithPosition = regexp.MustCompile(`- Error at: <(.*):([0-9]+):([0-9]+)> (.*)\n(.*)`)
var errorsBefore = regexp.MustCompile(`ERROR (.*) - (.*)`)

func ParseGobraOutput(output string) (VerificationResponse, error) {
	r := VerificationResponse{}
	r.Errors = make([]VerificationError, 0)
	if numErrorRe.FindString(output) == "" {
		for _, match := range errorsBefore.FindAllStringSubmatch(output, -1) {
			if len(match) < 2 {
				return r, errors.New("failed to parse error message")
			}
			r.Errors = append(r.Errors, VerificationError{
				Message:  match[2],
				Position: Position{0, 0},
			})
		}
	} else {
		for _, match := range errorWithPosition.FindAllStringSubmatch(output, -1) {
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
	}
	return r, nil
}
