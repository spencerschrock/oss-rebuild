// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package local

import (
	"context"
	"io"
	"sync"

	"github.com/google/oss-rebuild/pkg/build"
)

// LocalHandle implements build.Handle for local Docker builds
type LocalHandle struct {
	id         string
	cancel     context.CancelFunc
	output     io.ReadWriteCloser
	resultChan chan build.Result

	statusMu sync.RWMutex
	status   build.BuildState
}

// BuildID implements build.Handle
func (h *LocalHandle) BuildID() string {
	return h.id
}

// Wait implements build.Handle
func (h *LocalHandle) Wait(ctx context.Context) (build.Result, error) {
	defer h.output.Close()
	select {
	case result := <-h.resultChan:
		return result, nil
	case <-ctx.Done():
		// Context timeout - this is different from build cancellation
		return build.Result{}, ctx.Err()
	}
}

// OutputStream implements build.Handle
func (h *LocalHandle) OutputStream() io.Reader {
	return h.output
}

// Status implements build.Handle
func (h *LocalHandle) Status() build.BuildState {
	h.statusMu.RLock()
	defer h.statusMu.RUnlock()
	return h.status
}

// Cancel cancels the build
func (h *LocalHandle) Cancel() {
	defer h.output.Close()
	h.cancel()
}

// UpdateStatus updates the handle's status
func (h *LocalHandle) UpdateStatus(state build.BuildState) {
	h.statusMu.Lock()
	defer h.statusMu.Unlock()
	h.status = state
}

// SetResult sets the final result and closes the result channel
func (h *LocalHandle) SetResult(result build.Result) {
	select {
	case h.resultChan <- result:
	default:
		// Channel already closed or full
	}
}

// writeOutput writes a line to the output stream
func (h *LocalHandle) Write(line []byte) (n int, err error) {
	return h.output.Write(line)
}

// NewLocalHandle creates a new LocalHandle
func NewLocalHandle(id string, cancel context.CancelFunc, output io.ReadWriteCloser) *LocalHandle {
	return &LocalHandle{
		id:         id,
		cancel:     cancel,
		output:     output,
		resultChan: make(chan build.Result, 1),
		status:     build.BuildStateStarting,
	}
}
