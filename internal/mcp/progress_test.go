package mcp

import (
	"testing"
)

func TestProgressWriter_BatchesLines(t *testing.T) {
	var chunks []string
	pw := NewProgressWriter(3, func(chunk string, total int) {
		chunks = append(chunks, chunk)
	})

	pw.Write([]byte("line1\nline2\nline3\nline4\n"))
	pw.Flush()

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	// First chunk: 3 lines.
	if chunks[0] != "line1\nline2\nline3" {
		t.Errorf("chunk 0 = %q", chunks[0])
	}
	// Second chunk: 1 remaining line.
	if chunks[1] != "line4" {
		t.Errorf("chunk 1 = %q", chunks[1])
	}
}

func TestProgressWriter_PartialLines(t *testing.T) {
	var total int
	pw := NewProgressWriter(5, func(chunk string, t int) {
		total = t
	})

	pw.Write([]byte("partial"))
	pw.Write([]byte(" line\n"))
	pw.Flush()

	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
}

func TestProgressWriter_TotalLines(t *testing.T) {
	pw := NewProgressWriter(10, nil)
	pw.Write([]byte("a\nb\nc\n"))

	if pw.TotalLines() != 3 {
		t.Errorf("total = %d, want 3", pw.TotalLines())
	}
}
