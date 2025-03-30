package parser

import (
	"github.com/gottschali/gobra-playground/lib/util"
	"testing"
)

var tests = map[string]VerificationResponse{
	"tests/0-errors.txt": {
		Verified: true,
		Errors:   []VerificationError{},
	},
	"tests/1-errors.txt": {
		Verified: false,
		Errors: []VerificationError{
			{
				Message: `Precondition of call Abs(MinInt) might not hold.
Assertion x != MinInt might not hold.`,
				Position: Position{29, 8},
			},
		},
	},
	"tests/3-errors.txt": {
		Verified: false,
		Errors: []VerificationError{
			{
				Message: `Precondition of call foo(y) might not hold.
Assertion x > 0 might not hold.`,
				Position: Position{15, 2},
			},
			{
				Message: `Assert might fail.
Assertion x == y might not hold.`,
				Position: Position{11, 2},
			},
			{
				Message: `Postcondition might not hold.
Assertion res > 0 might not hold.`,
				Position: Position{5, 9},
			},
		},
	},
	"tests/violation.txt": {
		Verified: false,
		Errors: []VerificationError{
			{
				Message:  `An assumption was violated during execution.`,
				Position: Position{0, 0},
			},
			{
				Message:  `Logic error: Missing package clause in /tmp/gobra-playground3035386246/input.gobra`,
				Position: Position{0, 0},
			},
		},
	},
	"tests/logic-exception.txt": {
		Verified: false,
		Errors: []VerificationError{
			{
				Message:  `An assumption was violated during execution.`,
				Position: Position{0, 0},
			},
			{
				Message:  `Logic error: This case should be unreachable, but got unknown`,
				Position: Position{0, 0},
			},
		},
	},
}

func compare(v1, v2 VerificationResponse) bool {
	if v1.Verified != v2.Verified {
		return false
	}
	if v1.Timeout != v2.Timeout {
		return false
	}
	if len(v1.Errors) != len(v2.Errors) {
		return false
	}
	for i, e1 := range v1.Errors {
		e2 := v2.Errors[i]
		if e1.Message != e2.Message {
			return false
		}
		if e1.Position != e2.Position {
			return false
		}
	}
	return true
}

func TestErrorParser(t *testing.T) {
	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			out := util.ReadTest(path, t)
			ans, err := ParseGobraOutput(out)
			if err != nil {
				t.Errorf("unwanted error: %s", err)
			}
			if !compare(ans, want) {
				t.Errorf("got %v, want %v", ans, want)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	r, e := ParseGobraOutput(` 13:26:53.757 [ForkJoinPool-3-worker-3] ERROR viper.gobra.reporting.FileWriterReporter - Error at: </home/ali/Code/gobra-playground/tests/error/multiple.gobra:15:2> Precondition of call foo(y) might not hold.
Assertion x > 0 might not hold.
13:26:53.757 [ForkJoinPool-3-worker-2] ERROR viper.gobra.reporting.FileWriterReporter - Error at: </home/ali/Code/gobra-playground/tests/error/multiple.gobra:11:2> Assert might fail.
Assertion x == y might not hold.
13:26:53.758 [ForkJoinPool-3-worker-1] ERROR viper.gobra.reporting.FileWriterReporter - Error at: </home/ali/Code/gobra-playground/tests/error/multiple.gobra:5:9> Postcondition might not hold.
Assertion res > 0 might not hold.
`)
	if e != nil {
		t.Log("Error: ", e)
	}
	t.Log("parsed: ", r)
}

// func TestSingle(path string, expected VerificationResponse) {
// 	out := readTest(path, t)
// 	resp, err := ParseGobraOutput(out)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !resp.Verified {
// 		t.Fatal("Failed to parse that verification succeeded")
// 	}

// 	if resp.Errors.ErrorMessage != "" {
// 		t.Fatal("There is an error message when there should be none")
// 	}
// }

// func TestErrors0(t *testing.T) {
// }

// func TestErrors1(t *testing.T) {
// 	out := readTest("./tests/output/1-error.txt", t)
// 	resp, err := ParseGobraOutput(out)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if resp.Verified {
// 		t.Fatal("Failed to parse that verification failed")
// 	}

// 	if resp.Errors.ErrorMessage != "Assertion x != MinInt might not hold." {
// 		t.Fatalf("Failed to parse error message, got %s", resp.Errors.ErrorMessage)
// 	}
// }

// func TestErrors3(t *testing.T) {
// 	out := readTest("./tests/output/3-errors.txt", t)
// 	resp, err := ParseGobraOutput(out)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if resp.Verified {
// 		t.Fatal("Failed to parse that verification failed")
// 	}

// 	if resp.Errors.ErrorMessage != "Assertion x != MinInt might not hold." {
// 		t.Fatalf("Failed to parse error message, got %s", resp.Errors.ErrorMessage)
// 	}
// }
