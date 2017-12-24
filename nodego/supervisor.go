// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nodego

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

type logEntry struct {
	TextPayload string
	Severity    string
	Time        string
	ExecutionID string
}

func (e *logEntry) consoleOutput() []byte {
	var logBuf bytes.Buffer
	fmt.Fprintf(&logBuf, "[%s]", e.Severity[:1])
	fmt.Fprintf(&logBuf, "[%s]", time.Now().Format(isoTimeFormat))
	if e.ExecutionID != "" {
		fmt.Fprintf(&logBuf, "[%s]", e.ExecutionID)
	}
	logBuf.WriteByte(' ')
	logBuf.WriteString(e.TextPayload)
	if len(e.TextPayload) == 0 || e.TextPayload[len(e.TextPayload)-1] != '\n' {
		logBuf.WriteByte('\n')
	}
	return logBuf.Bytes()
}

type logBatch struct {
	Entries []*logEntry

	payloadLength int
	ready         chan struct{}
}

// addEntry adds a log entry to the batch.
//
// Note: addEntry is not thread safe.
func (batch *logBatch) addEntry(entry *logEntry) {
	if batch.Entries == nil {
		close(batch.ready)
	}

	batch.Entries = append(batch.Entries, entry)
	batch.payloadLength += len(entry.TextPayload)
}

func (batch *logBatch) report() error {
	if len(batch.Entries) == 0 {
		return nil
	}

	if err := postToSupervisor("/_ah/log", batch, supervisorLogTimeout); err != nil {
		return err
	}

	return nil
}

type loggingContext struct {
	initOnce sync.Once

	queueMutex   sync.Mutex
	queue        chan *logBatch
	currentBatch *logBatch

	execIDMutex sync.RWMutex
	execID      string
}

// startNewBatch prepares a new batch.
//
// Note: startNewBatch is not thread safe.
func (ctx *loggingContext) startNewBatch() *logBatch {
	ctx.currentBatch = &logBatch{
		ready: make(chan struct{}),
	}
	ctx.queue <- ctx.currentBatch
	return ctx.currentBatch
}

func (ctx *loggingContext) setExecutionID(id string) {
	ctx.execIDMutex.Lock()
	ctx.execID = id
	ctx.execIDMutex.Unlock()
}

func (ctx *loggingContext) executionID() string {
	ctx.execIDMutex.RLock()
	defer ctx.execIDMutex.RUnlock()
	return ctx.execID
}

func (ctx *loggingContext) addEntry(entry *logEntry) bool {
	if ctx.queue == nil {
		return false
	}

	ctx.queueMutex.Lock()
	defer ctx.queueMutex.Unlock()

	// Start a new batch if the current one would grow too much.
	if len(ctx.currentBatch.Entries) > 0 &&
		(len(ctx.currentBatch.Entries)+1 > maxLogBatchEntries ||
			ctx.currentBatch.payloadLength+len(entry.TextPayload) > maxLogBatchLength) {
		ctx.startNewBatch()
	}

	ctx.currentBatch.addEntry(entry)

	return true
}

func (ctx *loggingContext) startReportWorker() {
	for logBatch := range ctx.queue {
		<-logBatch.ready

		ctx.queueMutex.Lock()
		if logBatch == ctx.currentBatch {
			ctx.startNewBatch()
		}
		ctx.queueMutex.Unlock()

		if err := logBatch.report(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			killInstance()
		}
	}
}

func (ctx *loggingContext) initialize() {
	ctx.initOnce.Do(func() {
		ctx.queue = make(chan *logBatch, 5)
		ctx.startNewBatch()
		go ctx.startReportWorker()
	})
}

var loggingCtx loggingContext

var (
	infoLogWriter  supervisorWriter = "INFO"
	errorLogWriter supervisorWriter = "ERROR"

	// InfoLogger is a logger that batches sends logs to the supervisor with a
	// severity level of INFO.
	InfoLogger = log.New(infoLogWriter, "", 0)
	// InfoLogger is a logger that batches sends logs to the supervisor with a
	// severity level of ERROR.
	ErrorLogger = log.New(errorLogWriter, "", 0)
)

func init() {
	if supervisorHostname != "" && supervisorInternalPort != "" {
		loggingCtx.initialize()
	}
}

const isoTimeFormat = "2006-01-02T15:04:05.999Z07:00"

type supervisorWriter string

// Write implements io.Writer.Write.
func (w supervisorWriter) Write(p []byte) (int, error) {
	entry := &logEntry{
		TextPayload: string(p),
		Severity:    string(w),
		Time:        time.Now().Format(isoTimeFormat),
		ExecutionID: loggingCtx.executionID(),
	}

	if !loggingCtx.addEntry(entry) {
		return os.Stderr.Write(entry.consoleOutput())
	}

	return len(entry.TextPayload), nil
}

func newSupervisorRequest(path string, v interface{}) (*http.Request, error) {
	postData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", (&url.URL{
		Scheme: "http",
		Host:   supervisorHostname + ":" + supervisorInternalPort,
		Path:   path,
	}).String(), bytes.NewBuffer(postData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(postData)))

	return req, nil
}

func doRequestWithContext(ctx context.Context, r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(r.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
	}
	return resp, err
}

func postToSupervisor(path string, v interface{}, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := newSupervisorRequest(path, v)
	if err != nil {
		return err
	}

	resp, err := doRequestWithContext(ctx, req)
	if err == ctx.Err() {
		return errors.New("timeout when calling supervisor")
	} else if err != nil {
		return fmt.Errorf("error when calling supervisor: %s\n\n%s\n", err.Error(), debug.Stack())
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("incorrect response code from supervisor: %d\n", resp.StatusCode)
	}

	return err
}

func killInstance() {
	err := postToSupervisor("/_ah/kill", nil, supervisorKillTimeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	// Exit code 16 is copied over from worker.js.
	os.Exit(16)
}

func loggerMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggingCtx.setExecutionID(r.Header.Get("Function-Execution-Id"))

		defer func() {
			loggingCtx.setExecutionID("")
		}()

		handler.ServeHTTP(w, r)
	}
}

// WithLogger returns an http.Handler that reads the function execution ID,
// attaches it to log messages sent to the supervisor.
func WithLogger(handler http.Handler) http.Handler {
	return loggerMiddleware(handler)
}

// WithLoggerFunc is the same as WithLogger but accepts a handler function.
func WithLoggerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return loggerMiddleware(handler)
}

// OverrideLogger sets the default logger output to the supervisor logger with
// a severity level of INFO.
func OverrideLogger() {
	log.SetOutput(infoLogWriter)
	log.SetFlags(0)
}
