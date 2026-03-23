package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LogWriter struct {
	dir  string
	file *os.File
}

func NewLogWriter(dir string) *LogWriter {
	return &LogWriter{dir: dir}
}

func (w *LogWriter) Create(catalogName, runbookName string, t time.Time) (string, error) {
	filename := fmt.Sprintf("%s-%s-%s.log",
		t.Format("2006.01.02-150405"),
		catalogName,
		runbookName,
	)
	path := filepath.Join(w.dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create log file: %w", err)
	}

	w.file = f
	return path, nil
}

func (w *LogWriter) WriteLine(line string) {
	if w.file != nil {
		fmt.Fprintln(w.file, line)
	}
}

func (w *LogWriter) Close() {
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}
}
