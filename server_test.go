package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gottschali/gobra-playground/lib/parser"
	"github.com/gottschali/gobra-playground/lib/util"
)

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
	if !strings.Contains(util.SafeString(body), "Gobra") {
		t.Errorf("response does not contain Gobra")
	}
}

type VerificationServer struct {
	server *httptest.Server
}

func MakeServer() VerificationServer {
	r := VerificationServer{httptest.NewServer(http.HandlerFunc(Verify))}
	return r
}

func (s VerificationServer) submit(code string) (*parser.VerificationResponse, error) {
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
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	parsed := new(parser.VerificationResponse)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
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
	code := util.ReadTest(path, t)
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
	code := util.ReadTest(path, t)
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

	data := url.Values{}
	data.Set("version", "1.0")
	data.Set("body", "package main\nassert true")
	r, _ := http.NewRequest(
		"POST",
		server.URL,
		strings.NewReader(data.Encode()),
	)
	// No Content-Type header
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
