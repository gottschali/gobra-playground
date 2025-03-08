package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/tools/playground"
)

// healthcheck
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

// type process struct {
// 	out  chan<- *Message
// 	done chan struct{} // closed when wait completes
// 	run  *exec.Cmd
// 	path string
// }

type defaultWriter struct{}

// safeString returns b as a valid UTF-8 string.
func safeString(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	var buf bytes.Buffer
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]
		buf.WriteRune(r)
	}
	return buf.String()
}

func (d defaultWriter) Write(s []byte) (n int, err error) {
	fmt.Println(safeString(s))
	return cap(s), nil
}

// IDEA: goify before sending to playground

// Note: go playground uses a websocket connection
// - allow to cancel

// TODO I think there is only the start in the error message
// for the exact position we need to look at stat.json?
type Position struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine"`
	StartPos  int `json:"startPos"`
	EndPos    int `json:"endPos"`
}

type VerificationError struct {
	ErrorMessage string   `json:"errorMessage"`
	Position     Position `json:"Position"`
}

// where stats
type VerificationResponse struct {
	Verified  bool              `json:"verified"`
	MainError VerificationError `json:"mainError"`
	// output: list[error string] `json:"output"`
	Duration float64 `json:"duration"`
	Stats    string  `json:"stats"`
}

var dev = os.Getenv("DEV")
var gobra_path string
var java_path = "java"

func verify(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

	// https://stackoverflow.com/questions/15407719/in-gos-http-package-how-do-i-get-the-query-string-on-a-post-request
	req.ParseForm()
	body := req.Form.Get("body")
	// TODO: allow more options

	fmt.Println(body)

	// write to a temporary file
	// https://gobyexample.com/temporary-files-and-directories
	temp_dir, _ := os.MkdirTemp("", "sampledir")
	// remove it afterwards (or maybe allow caching)

	// https://gobyexample.com/writing-files
	input_path := temp_dir + "/input.gobra"
	_ = os.WriteFile(input_path, []byte(body), 0644)

	// TODO separate  java args and gobra args
	// see the gobra action docker entrypoint for options
	// TODO time it
	args := []string{"java", "-jar", "-Xss128m", gobra_path, "--input", input_path, "-g", temp_dir}

	done := make(chan int)

	var outbuf, errbuf strings.Builder

	cmd := &exec.Cmd{
		Path:   java_path,
		Args:   args,
		Stdin:  strings.NewReader(""), // or just reader from empty string
		Stdout: &outbuf,
		Stderr: &errbuf,
	}

	go func() {
		// TODO proper error responses
		start := time.Now()
		if err := cmd.Start(); err != nil {
			fmt.Println("Error starting command", err)
			return
		}

		// fmt.Println(cmd.Env, gobra_path, java_path)

		if err := cmd.Wait(); err != nil {
			fmt.Println("Error waiting for the command:", err)
			return
		}
		elapsed := time.Since(start)

		dat, err := os.ReadFile(temp_dir + "/stats.json")
		if err != nil {
			fmt.Println("Failed to read stats.json", err)
			return
		}
		stats := safeString(dat)
		stdout := outbuf.String()
		stderr := errbuf.String()

		// TODO just basic logging
		fmt.Println("stats:", stats)
		fmt.Println("stdout:\n", stdout)
		fmt.Println("stderr:\n", stderr)

		resp, err := ParseGobraOutput(stdout)
		resp.Duration = elapsed.Seconds()
		if err != nil {
			fmt.Errorf("Error parsing output: %e", err)
		}
		fmt.Println(resp)
		resp.Stats = stats

		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Errorf("Error marshalling the response to json: %e", err)
		}
		w.Write(data)
		// fmt.Fprintf(w, string(data))

		defer os.RemoveAll(temp_dir)
		done <- 1
		close(done)
	}()
	// <-done
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("timed out")
		cmd.Process.Kill() // TODO
	case <-done:
		fmt.Println("Goroutine is done")
	}

}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer())))

	if dev != "" {
		gobra_path = "/home/ali/Code/gobra.jar"
		java_path = "/usr/bin/java"
	} else {
		gobra_path = "/gobra/gobra.jar"
	}

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/verify", verify)
	http.HandleFunc("/headers", headers)
	http.Handle("/compile2", playground.Proxy())

	PORT := ":8090"
	fmt.Println("Running on http://localhost", PORT)
	http.ListenAndServe(PORT, nil)
}
