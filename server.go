package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/gottschali/gobra-playground/lib/parser"
	"github.com/gottschali/gobra-playground/lib/util"
	"golang.org/x/tools/playground"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var gobra_jar = os.Getenv("GOBRA_JAR")
var java_exe = cmp.Or(os.Getenv("JAVA_EXE"), "java")
var port = cmp.Or(os.Getenv("PORT"), "8090")

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
	stats := util.SafeString(dat)

	resp, err := parser.ParseGobraOutput(util.SafeString(stdout))
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
	if _, err := w.Write(data); err != nil {
		errors <- err
	}
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

	java_args := []string{java_exe, "-jar", "-Xss128m"}
	gobra_args := []string{gobra_jar, "--input", input_path, "--gobraDirectory", dir}
	args := append(java_args, gobra_args...)

	cmd := &exec.Cmd{
		Path:  java_exe,
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
		if err := cmd.Process.Kill(); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		resp := parser.VerificationResponse{
			Timeout:  true,
			Duration: timeout.Seconds(),
		}
		data, _ := json.Marshal(resp)
		_, err := w.Write(data)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
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

// middleware to add cors headers
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func redirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = "/compile"
		next.ServeHTTP(w, req)
	})

}
func start() {
	http.Handle("/hello", cors(http.HandlerFunc(Hello)))
	http.Handle("/verify", cors(http.HandlerFunc(Verify)))
	http.Handle("/run", cors(redirect(playground.Proxy())))

	fmt.Println("Starting server on http://localhost:", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(`ERROR: Failed to start the server.
Check if the port is already in use.`)
	}
}

func init() {
	if gobra_jar == "" {
		fmt.Println("ERROR: GOBRA_JAR environment variable must be set")
		os.Exit(1)
	}
}

func main() {
	start()
}
