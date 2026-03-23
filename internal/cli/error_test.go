package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatError(t *testing.T) {
	var buf bytes.Buffer
	FormatError(&buf, "Runbook not found", "runbook \"unknown.id\" not found")

	out := buf.String()

	if !strings.Contains(out, "ERROR") {
		t.Error("output should contain ERROR badge")
	}
	if !strings.Contains(out, "Runbook not found") {
		t.Error("output should contain title")
	}
	if !strings.Contains(out, "runbook \"unknown.id\" not found") {
		t.Error("output should contain detail")
	}
}

func TestFormatError_NoDetail(t *testing.T) {
	var buf bytes.Buffer
	FormatError(&buf, "Something failed", "")

	out := buf.String()

	if !strings.Contains(out, "ERROR") {
		t.Error("output should contain ERROR badge")
	}
	if !strings.Contains(out, "Something failed") {
		t.Error("output should contain title")
	}
}
