// buildhost is a generic build broker. It accepts a source tarball + a script
// path, queues the job, and streams the script's output back to the client.
// All build logic lives in the scripts/ directory; this server is agnostic.
//
// Endpoints:
//   GET  /              — bootstrap script (run on the Mac via `curl … | bash`)
//   POST /build?script= — Linux client uploads source, streams output back
//   GET  /next          — runner long-polls, gets `run_job '<id>' '<script>'`
//   GET  /source/<id>   — runner downloads source tarball
//   POST /log/<id>      — runner streams build output here (forwarded to /build)
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type job struct {
	id     string
	script string
	source []byte

	// Pipe: /log/<id> writes to logW, /build reads from logR.
	logR *io.PipeReader
	logW *io.PipeWriter
}

type server struct {
	queue chan *job
	mu    sync.Mutex
	byID  map[string]*job
}

func newServer() *server {
	return &server{
		queue: make(chan *job),
		byID:  map[string]*job{},
	}
}

func (s *server) register(j *job) {
	s.mu.Lock()
	s.byID[j.id] = j
	s.mu.Unlock()
}

func (s *server) unregister(id string) {
	s.mu.Lock()
	delete(s.byID, id)
	s.mu.Unlock()
}

func (s *server) lookup(id string) *job {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byID[id]
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// validScriptPath blocks path traversal and absolute paths; the runner trusts
// the value enough to pass it through bash, so we keep it conservative.
func validScriptPath(p string) bool {
	if p == "" || len(p) > 256 {
		return false
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "..") {
		return false
	}
	for _, c := range p {
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '.', c == '/', c == '_', c == '-':
		default:
			return false
		}
	}
	return true
}

// GET / — bootstrap. The Mac fetches this once per session.
func (s *server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	fmt.Fprintf(w, bootstrapScript, r.Host)
}

// POST /build?script=<path> — Linux uploads source, blocks while streaming.
func (s *server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	script := r.URL.Query().Get("script")
	if !validScriptPath(script) {
		http.Error(w, "invalid script path", http.StatusBadRequest)
		return
	}
	source, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	j := &job{
		id:     newID(),
		script: script,
		source: source,
	}
	j.logR, j.logW = io.Pipe()
	s.register(j)
	defer s.unregister(j.id)

	log.Printf("[%s] queued script=%s (%d bytes source)", j.id, j.script, len(source))

	select {
	case s.queue <- j:
	case <-r.Context().Done():
		log.Printf("[%s] client cancelled before pickup", j.id)
		return
	}

	// Cancel the pipe if the client disconnects mid-stream.
	go func() {
		<-r.Context().Done()
		j.logR.CloseWithError(r.Context().Err())
	}()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)

	buf := make([]byte, 4096)
	for {
		n, err := j.logR.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				log.Printf("[%s] client write error: %v", j.id, werr)
				j.logR.CloseWithError(werr)
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] log read ended: %v", j.id, err)
			}
			break
		}
	}
	log.Printf("[%s] streaming done", j.id)
}

// GET /next — runner long-polls; returns a single-line `run_job` invocation.
func (s *server) handleNext(w http.ResponseWriter, r *http.Request) {
	var j *job
	select {
	case j = <-s.queue:
	case <-r.Context().Done():
		return
	}
	log.Printf("[%s] picked up by runner %s (script=%s)", j.id, r.RemoteAddr, j.script)
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	// run_job is defined and exported in the bootstrap, so this single
	// function call is enough to kick off the full job.
	fmt.Fprintf(w, "run_job '%s' '%s'\n", j.id, j.script)
}

// GET /source/<id>
func (s *server) handleSource(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/source/")
	j := s.lookup(id)
	if j == nil {
		http.Error(w, "no such job", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(j.source)
}

// POST /log/<id> — runner streams build output here.
func (s *server) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "POST or PUT required", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/log/")
	j := s.lookup(id)
	if j == nil {
		http.Error(w, "no such job", http.StatusNotFound)
		return
	}
	_, err := io.Copy(j.logW, r.Body)
	if err != nil {
		log.Printf("[%s] log upload error: %v", j.id, err)
		j.logW.CloseWithError(err)
	} else {
		j.logW.Close() // EOF for /build reader
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	s := newServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/build", s.handleBuild)
	mux.HandleFunc("/next", s.handleNext)
	mux.HandleFunc("/source/", s.handleSource)
	mux.HandleFunc("/log/", s.handleLog)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}
	addr := ":" + port
	log.Printf("buildhost listening on %s", addr)
	log.Printf("on the Mac: curl localhost:%s | bash", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// The bootstrap is fetched ONCE per runner session. It defines run_job (the
// generic source-download + script-runner + log-streamer + status-reporter)
// and exports it so each /next response — which is just a one-line call to
// run_job — can use it in the bash subshell that `curl /next | bash` spawns.
//
// Changes to the build commands themselves do NOT require restarting the
// runner: those live in scripts/*.sh inside the source tarball, hot-reloaded
// on every job. Only changes to run_job or prereq checks need a runner restart.
const bootstrapScript = `#!/usr/bin/env bash
set -uo pipefail
SERVER="http://%s"
export SERVER

echo "=== buildhost runner ==="

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: missing '$1'."
    echo "  Install: $2"
    exit 1
  fi
}

soft_need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "WARN: '$1' not found. $2"
  fi
}

need go       "brew install go"
need xcrun    "xcode-select --install (or install Xcode from the App Store)"
need gomobile "go install golang.org/x/mobile/cmd/gomobile@latest"
need gobind   "go install golang.org/x/mobile/cmd/gobind@latest"
soft_need xcodegen "iOS app builds need it. Install: brew install xcodegen"

# run_job: invoked by each /next response. Downloads source, runs the
# requested script, streams combined stdout/stderr to the broker, and
# appends a __BUILD_STATUS marker so the client knows the outcome.
run_job() {
  local JOB_ID="$1" SCRIPT="$2"
  local WORK
  WORK=$(mktemp -d -t buildhost-job-XXXXXX)
  # shellcheck disable=SC2064
  trap "rm -rf '$WORK'" RETURN

  echo "[$JOB_ID] starting (script=$SCRIPT)" >&2

  {
    {
      set -e
      cd "$WORK"
      echo "==> downloading source"
      curl -sSf "$SERVER/source/$JOB_ID" -o source.tar.gz
      mkdir src && tar xzf source.tar.gz -C src
      cd src
      echo "==> running $SCRIPT"
      bash "$SCRIPT"
    } 2>&1
    local RC=$?
    if [ "$RC" -eq 0 ]; then
      echo "__BUILD_STATUS:ok__"
    else
      echo "__BUILD_STATUS:fail__ (exit $RC)"
    fi
  } | curl -sS -X POST -T - "$SERVER/log/$JOB_ID" || true

  echo "[$JOB_ID] done" >&2
}
export -f run_job

echo "Prereqs OK. Polling $SERVER for jobs (Ctrl-C to stop)..."

while true; do
  if ! curl -sNf "$SERVER/next" | bash; then
    # Server gone or job failed; back off briefly so we don't hot-loop.
    sleep 2
  fi
done
`
