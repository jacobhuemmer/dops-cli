package mcp

import (
	"strings"
	"sync"
)

// defaultProgressBatchSize is the number of complete lines to accumulate before
// flushing a progress notification.
const defaultProgressBatchSize = 5

// ProgressCallback is called with a batch of output lines.
type ProgressCallback func(chunk string, linesSoFar int)

// ProgressWriter is an io.Writer that accumulates output, batches complete
// lines, and calls a callback for progress notifications.
type ProgressWriter struct {
	mu        sync.Mutex
	buf       strings.Builder
	lines     []string
	batchSize int
	callback  ProgressCallback
	total     int
}

// NewProgressWriter creates a progress writer that batches lines and calls
// the callback every batchSize complete lines.
func NewProgressWriter(batchSize int, callback ProgressCallback) *ProgressWriter {
	if batchSize <= 0 {
		batchSize = defaultProgressBatchSize
	}
	return &ProgressWriter{
		batchSize: batchSize,
		callback:  callback,
	}
}

func (w *ProgressWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Write(p)

	// Extract complete lines.
	content := w.buf.String()
	for {
		idx := strings.IndexByte(content, '\n')
		if idx < 0 {
			break
		}
		line := content[:idx]
		content = content[idx+1:]
		w.lines = append(w.lines, line)
		w.total++

		if len(w.lines) >= w.batchSize {
			w.flush()
		}
	}

	w.buf.Reset()
	w.buf.WriteString(content)
	return len(p), nil
}

// Flush sends any remaining buffered lines as a final progress notification.
func (w *ProgressWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Flush any incomplete line in buffer.
	remaining := w.buf.String()
	if remaining != "" {
		w.lines = append(w.lines, remaining)
		w.total++
		w.buf.Reset()
	}

	if len(w.lines) > 0 {
		w.flush()
	}
}

func (w *ProgressWriter) flush() {
	if w.callback != nil && len(w.lines) > 0 {
		chunk := strings.Join(w.lines, "\n")
		w.callback(chunk, w.total)
	}
	w.lines = w.lines[:0]
}

// TotalLines returns the total number of lines written so far.
func (w *ProgressWriter) TotalLines() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.total
}
