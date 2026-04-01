package adapters

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "bare tilde",
			path: "~",
			want: home,
		},
		{
			name: "tilde with forward slash",
			path: "~/documents/file.txt",
			want: filepath.Join(home, "documents/file.txt"),
		},
		{
			name: "no tilde",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "tilde in middle is not expanded",
			path: "/foo/~/bar",
			want: "/foo/~/bar",
		},
		{
			name: "empty string",
			path: "",
			want: "",
		},
		{
			name: "tilde only prefix no sep",
			path: "~username",
			want: "~username",
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name string
			path string
			want string
		}{
			name: "tilde with backslash",
			path: `~\documents\file.txt`,
			want: filepath.Join(home, `documents\file.txt`),
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandHome(tt.path)
			if got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestOSFileSystem_ReadWriteFile(t *testing.T) {
	fs := NewOSFileSystem()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := fs.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	data, err := fs.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("got %q, want hello", data)
	}
}

func TestOSFileSystem_ReadDir(t *testing.T) {
	fs := NewOSFileSystem()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)

	entries, err := fs.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("entries = %d, want 2", len(entries))
	}
}

func TestOSFileSystem_MkdirAll(t *testing.T) {
	fs := NewOSFileSystem()
	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	if err := fs.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
}

func TestOSFileSystem_Stat(t *testing.T) {
	fs := NewOSFileSystem()
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	os.WriteFile(path, []byte("x"), 0o644)

	info, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.IsDir() {
		t.Error("should not be a directory")
	}

	_, err = fs.Stat(filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Error("Stat nonexistent should error")
	}
}

func TestOSFileSystem_ReadFile_Error(t *testing.T) {
	fs := NewOSFileSystem()
	_, err := fs.ReadFile("/nonexistent/path/xyz")
	if err == nil {
		t.Error("ReadFile nonexistent should error")
	}
}
