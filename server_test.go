package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

var successOutput = `Gobra 1.1-SNAPSHOT (46a35ffb@(detached))
(c) Copyright ETH Zurich 2012 - 2024
14:21:12.006 [main] INFO viper.gobra.Gobra - Verifying package /home/anonymous/Code/gobra-book-monorepo/scratch - bank [14:21:12]
14:21:15.583 [thread-5] INFO viper.gobra.Gobra - Gobra found no errors
14:21:15.583 [main] INFO viper.gobra.Gobra - Gobra has found 0 error(s)
14:21:15.604 [Thread-0] INFO viper.gobra.Gobra - Writing report to .gobra/stats.json`

var failedOutput = `Gobra 1.1-SNAPSHOT (46a35ffb@(detached))
(c) Copyright ETH Zurich 2012 - 2024
14:50:32.055 [main] INFO viper.gobra.Gobra - Verifying package /home/anonymous/Code/gobra-book-monorepo/scratch - main [14:50:32]
14:50:35.387 [ForkJoinPool-3-worker-2] ERROR viper.gobra.reporting.FileWriterReporter - Error at: <abs.go:29:8> Precondition of call Abs(MinInt) might not hold.
Assertion x != MinInt might not hold.
14:50:35.403 [thread-2] ERROR viper.gobra.Gobra - Gobra has found 1 error(s) in package /home/anonymous/Code/gobra-book-monorepo/scratch - main
14:50:35.403 [main] INFO viper.gobra.Gobra - Gobra has found 1 error(s)
14:50:35.427 [Thread-0] INFO viper.gobra.Gobra - Writing report to .gobra/stats.json
java -jar -Xss1g /home/anonymous/Code/gobra.jar -i abs.go`

func TestVerified(t *testing.T) {
	resp, err := ParseGobraOutput(successOutput)

	if err != nil {
		t.Fatal(err)
	}

	if !resp.Verified {
		t.Fatal("Failed to parse that verification succeeded")
	}

	if resp.MainError.ErrorMessage != "" {
		t.Fatal("There is an error message when there should be none")
	}
}

func TestFailDetected(t *testing.T) {
	resp, err := ParseGobraOutput(failedOutput)

	if err != nil {
		t.Fatal(err)
	}

	if resp.Verified {
		t.Fatal("Failed to parse that verification failed")
	}

	if resp.MainError.ErrorMessage != "Assertion x != MinInt might not hold." {
		t.Fatalf("Failed to parse error message, got %s", resp.MainError.ErrorMessage)
	}
}

func TestHealthcheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(Hello))
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("Healthcheck failed: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(safeString(body), "Gobra") {
		t.Errorf("response does not contain Gobra")
	}
}

func readTest(path string, t *testing.T) string {
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Test file does not exist: %s", err)
	}
	return safeString(contents)
}

type VerificationServer struct {
	server *httptest.Server
}

func MakeServer() VerificationServer {
	r := VerificationServer{httptest.NewServer(http.HandlerFunc(Verify))}
	return r
}

func (s VerificationServer) submit(code string) (*VerificationResponse, error) {
	data := url.Values{}
	data.Set("version", "1.0")
	data.Set("body", code)
	r, _ := http.NewRequest(
		"POST",
		s.server.URL,
		strings.NewReader(data.Encode()),
	)
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.server.Client().Do(r)

	defer resp.Body.Close()
	parsed := new(VerificationResponse)
	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func TestVerifies(t *testing.T) {
	s := MakeServer()
	defer s.server.Close()
	path := "./tests/tutorial/basicAnnotations.gobra"
	code := readTest(path, t)
	resp, err := s.submit(code)
	if err != nil {
		t.Fatalf("error submitting code: %s", err)
	}
	if !resp.Verified {
		t.Errorf("Wrong response: test should have verified: %s", path)
	}
}

func TestVerifiesFail(t *testing.T) {
	s := MakeServer()
	defer s.server.Close()
	path := "./tests/error/array-length-fail2.gobra"
	code := readTest(path, t)
	resp, err := s.submit(code)
	if err != nil {
		t.Fatalf("error submitting code: %s", err)
	}

	if resp.Verified {
		t.Errorf("Wrong response: test should not have verified")
	}

}

func TestNoContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(Verify))
	defer server.Close()

	// required field body is missing
	data := url.Values{}
	data.Set("version", "1.0")
	data.Set("body", "package main\nassert true")
	r, _ := http.NewRequest(
		"POST",
		server.URL,
		strings.NewReader(data.Encode()),
	)
	r.Header.Add("Accept", "application/json")
	resp, _ := server.Client().Do(r)

	if resp.StatusCode < 400 {
		t.Fatalf("no error code when data is not urlencoded ")
	}

}

func TestMissingBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(Verify))
	defer server.Close()

	// required field body is missing
	data := url.Values{}
	data.Set("version", "1.0")
	r, _ := http.NewRequest(
		"POST",
		server.URL,
		strings.NewReader(data.Encode()),
	)
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := server.Client().Do(r)

	if resp.StatusCode < 400 {
		t.Fatalf("no error code when field body is missing")
	}

}
