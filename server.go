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

// returns b as a valid UTF-8 string.
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
	Timeout   bool              `json:"timeout"`
	MainError VerificationError `json:"mainError"`
	// output: list[error string] `json:"output"`
	Duration float64 `json:"duration"`
	Stats    string  `json:"stats"`
}

var dev = os.Getenv("DEV") != ""
var gobra_path = os.Getenv("GOBRA_PATH")
var java_path = os.Getenv("JAVA_PATH")
var port = os.Getenv("PORT")

func Verify(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	// https://stackoverflow.com/questions/15407719/in-gos-http-package-how-do-i-get-the-query-string-on-a-post-request
	err := req.ParseForm()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	body := req.Form.Get("body")
	// TODO: allow more options
	// probably as json in options form field
	// For gobra consider
	// --overflow
	// --eraseGhost
	// --backend
	// See java -jar gobra.jar -h
	// or the gobra ci action for args
	// req.Form.Get("version")
	// req.Form.Get("options")

	fmt.Println(req.URL)
	if dev {
		fmt.Println(body)
	}

	// write to a temporary file
	// https://gobyexample.com/temporary-files-and-directories
	temp_dir, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		fmt.Println("Failed to create a temporary directory.")
	}
	defer os.RemoveAll(temp_dir)

	// https://gobyexample.com/writing-files
	input_path := temp_dir + "/input.gobra"
	err = os.WriteFile(input_path, []byte(body), 0644)
	if err != nil {
		fmt.Println("failed to write gobra file", input_path)
		http.Error(w, "internal error", 500)
		return
	}

	java_args := []string{java_path, "-jar", "-Xss128m"}
	gobra_args := []string{gobra_path, "--input", input_path, "-g", temp_dir}
	args := append(java_args, gobra_args...)

	done := make(chan int)
	errors := make(chan error)

	var outbuf, errbuf strings.Builder

	cmd := &exec.Cmd{
		Path:   java_path,
		Args:   args,
		Stdin:  strings.NewReader(""), // or just reader from empty string
		Stdout: &outbuf,
		Stderr: &errbuf,
	}

	go func() {
		start := time.Now()

		if err := cmd.Start(); err != nil {
			errors <- fmt.Errorf("Error starting command: %s", err)
			return
		}
		cmd.Wait()
		elapsed := time.Since(start)

		dat, err := os.ReadFile(temp_dir + "/stats.json")
		if err != nil {
			errors <- fmt.Errorf("Failed to read stats.json, %s", err)
			return
		}
		stats := safeString(dat)
		stdout := outbuf.String()
		stderr := errbuf.String()

		if dev {
			fmt.Println("stats:", stats)
			fmt.Println("stdout:\n", stdout)
			fmt.Println("stderr:\n", stderr)
		}

		resp, err := ParseGobraOutput(stdout)
		resp.Duration = elapsed.Seconds()
		if err != nil {
			errors <- fmt.Errorf("Error parsing output: %e", err)
			return
		}
		if dev {
			fmt.Println(resp)
		}
		resp.Stats = stats

		data, err := json.Marshal(resp)
		if err != nil {
			errors <- fmt.Errorf("Error marshalling the response to json: %e", err)
			return
		}
		w.Write(data)
		done <- 1
	}()
	const timeout = 10 * time.Second
	select {
	case <-time.After(timeout):
		fmt.Println("timed out")
		cmd.Process.Kill() // TODO
		resp := VerificationResponse{
			Timeout:  true,
			Duration: timeout.Seconds(),
		}
		data, _ := json.Marshal(resp)
		w.Write(data)
	case err := <-errors:
		fmt.Printf("internal error: %s\n", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	case <-done:
	}

}

// healthcheck
func Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Welcome to the Gobra Playground\n")
}

func start() {
	http.HandleFunc("/hello", Hello)
	http.HandleFunc("/verify", Verify)
	playground.Proxy() // /compile
	// http.Handle("/compile2", playground.Proxy())

	fmt.Println("Starting server on http://localhost:", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(`ERROR: Failed to start the server.
Check if the port is already in use.`)
	}
}

func main() {
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer())))

	if java_path == "" {
		fmt.Println("ERROR: JAVA_PATH environment variable must be set")
		return
	}
	if gobra_path == "" {
		fmt.Println("ERROR: GOBRA_PATH environment variable must be set")
		return
	}
	if dev {
		fmt.Println("Running in dev mode")
		fmt.Printf("JAVA_PATH=%s\n", java_path)
		fmt.Printf("GOBRA_PATH=%s\n", gobra_path)
	}
	if port == "" {
		port = "8090"
	}
	start()
}
