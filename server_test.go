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

func TestTable(t *testing.T) {
	var tests = []struct {
		path     string
		verifies bool
	}{
		{"tests/basicAnnotations.gobra", true},
		{"tests/array-length-fail2.gobra", false},
		{"tests/no-package.gobra", false},
		{"tests/logic-exception.gobra", false},
		{"tests/empty.gobra", false},
		{"tests/abs.gobra", true},
		{"tests/absFail.gobra", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			s := MakeServer()
			defer s.server.Close()
			code := util.ReadTest(tt.path, t)
			resp, err := s.submit(code)
			if err != nil {
				t.Fatalf("error submitting code: %s", err)
			}
			if resp.Verified != tt.verifies {
				t.Errorf("Wrong verification verdict: %s", tt.path)
			}
		})
	}

}
