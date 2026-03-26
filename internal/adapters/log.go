package adapters

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogWriter struct {
	dir    string
	file   *os.File
	writer *bufio.Writer
}

func NewLogWriter(dir string) *LogWriter {
	return &LogWriter{dir: dir}
}

func (w *LogWriter) Create(catalogName, runbookName string, t time.Time) (string, error) {
	if err := os.MkdirAll(w.dir, 0o700); err != nil {
		return "", fmt.Errorf("create log dir: %w", err)
	}

	filename := fmt.Sprintf("%s-%s-%s.log",
		t.Format("2006.01.02-150405"),
		catalogName,
		runbookName,
	)

	// Path traversal protection.
	baseClean := filepath.Clean(w.dir)
	targetClean := filepath.Clean(filepath.Join(baseClean, filename))
	rel, err := filepath.Rel(baseClean, targetClean)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid log filename: %q", filename)
	}

	f, err := os.Create(targetClean)
	if err != nil {
		return "", fmt.Errorf("create log file: %w", err)
	}

	w.file = f
	w.writer = bufio.NewWriterSize(f, 64*1024) // 64KB buffer
	return targetClean, nil
}

func (w *LogWriter) WriteLine(line string) {
	if w.writer != nil {
		// Best-effort; the TUI can't surface write errors during streaming.
		_, _ = w.writer.WriteString(line + "\n")
	}
}

func (w *LogWriter) Close() {
	if w.writer != nil {
		_ = w.writer.Flush()
		w.writer = nil
	}
	if w.file != nil {
		_ = w.file.Close() // best-effort cleanup
		w.file = nil
	}
}
