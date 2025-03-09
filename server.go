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

func gobra(w http.ResponseWriter, cmd *exec.Cmd, errors chan error, done chan int) {
	start := time.Now()
	stdout, err := cmd.Output()
	if err != nil && cmd.ProcessState.ExitCode() == 0 {
		errors <- fmt.Errorf("error running gobra process: %s", err)
		return
	}
	elapsed := time.Since(start)
	temp_dir := cmd.Args[len(cmd.Args)-1]
	dat, err := os.ReadFile(temp_dir + "/stats.json")
	if err != nil {
		errors <- fmt.Errorf("Failed to read stats.json, %s", err)
		return
	}
	stats := safeString(dat)

	resp, err := ParseGobraOutput(safeString(stdout))
	if err != nil {
		errors <- fmt.Errorf("Error parsing output: %e", err)
		return
	}
	resp.Duration = elapsed.Seconds()
	resp.Stats = stats
	if dev {
		fmt.Println(resp)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		errors <- fmt.Errorf("Error marshalling the response to json: %e", err)
		return
	}
	w.Write(data)
	done <- 1
}

func buildCommand(req *http.Request, dir string) (*exec.Cmd, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("Failed to parse form: %s", err)
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

	if dev {
		fmt.Println(body)
	}

	// https://gobyexample.com/writing-files
	input_path := dir + "/input.gobra"
	err = os.WriteFile(input_path, []byte(body), 0644)
	if err != nil {
		fmt.Println("failed to write gobra file", input_path)
		return nil, err
	}

	java_args := []string{java_path, "-jar", "-Xss128m"}
	gobra_args := []string{gobra_path, "--input", input_path, "-g", dir}
	args := append(java_args, gobra_args...)

	cmd := &exec.Cmd{
		Path:  java_path,
		Args:  args,
		Stdin: strings.NewReader(""), // or just reader from empty string
	}
	return cmd, nil
}

func Verify(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Method, req.URL, req.Host)
	defer req.Body.Close()

	temp_dir, err := os.MkdirTemp("", "sampledir")
	if err != nil {
		fmt.Println("Failed to create a temporary directory.")
		http.Error(w, "internal error", 500)
		return
	}
	defer os.RemoveAll(temp_dir)
	cmd, err := buildCommand(req, temp_dir)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	done := make(chan int)
	errors := make(chan error)

	go gobra(w, cmd, errors, done)
	const timeout = 10 * time.Second
	select {
	case <-time.After(timeout):
		fmt.Println("timed out")
		cmd.Process.Kill()
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
