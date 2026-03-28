package theme

import (
	"io/fs"
	"os"
	"testing"
)

type fakeFS struct {
	files map[string][]byte
}

func newFakeFS() *fakeFS {
	return &fakeFS{files: make(map[string][]byte)}
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeFS) WriteFile(string, []byte, fs.FileMode) error  { return nil }
func (f *fakeFS) ReadDir(string) ([]os.DirEntry, error)        { return nil, nil }
func (f *fakeFS) MkdirAll(string, fs.FileMode) error           { return nil }
func (f *fakeFS) Stat(string) (os.FileInfo, error)             { return nil, os.ErrNotExist }

func TestFileThemeLoader_BundledTheme(t *testing.T) {
	ffs := newFakeFS()
	loader := NewFileLoader(ffs, "/fake/themes")

	tf, err := loader.Load("github")
	if err != nil {
		t.Fatalf("Load bundled: %v", err)
	}

	if tf.Name != "github" {
		t.Errorf("name = %q, want github", tf.Name)
	}
	if _, ok := tf.Defs["blue"]; !ok {
		t.Error("expected 'blue' in defs")
	}
}

func TestFileThemeLoader_UserOverride(t *testing.T) {
	ffs := newFakeFS()
	ffs.files["/fake/themes/tokyonight.json"] = []byte(`{
		"name": "custom-tokyonight",
		"defs": {"bg": "#000000"},
		"theme": {"background": {"dark": "bg", "light": "bg"}}
	}`)
	loader := NewFileLoader(ffs, "/fake/themes")

	tf, err := loader.Load("tokyonight")
	if err != nil {
		t.Fatalf("Load user override: %v", err)
	}

	if tf.Name != "custom-tokyonight" {
		t.Errorf("name = %q, want custom-tokyonight (user override)", tf.Name)
	}
}

func TestFileThemeLoader_UnknownFallsBack(t *testing.T) {
	ffs := newFakeFS()
	loader := NewFileLoader(ffs, "/fake/themes")

	tf, err := loader.Load("nonexistent-theme")
	if err != nil {
		t.Fatalf("Load fallback: %v", err)
	}

	if tf.Name != "github" {
		t.Errorf("name = %q, want github (fallback)", tf.Name)
	}
}

func TestFileThemeLoader_UserThemeNotOverridingBundled(t *testing.T) {
	ffs := newFakeFS()
	ffs.files["/fake/themes/dracula.json"] = []byte(`{
		"name": "dracula",
		"defs": {"bg": "#282a36"},
		"theme": {"background": {"dark": "bg", "light": "bg"}}
	}`)
	loader := NewFileLoader(ffs, "/fake/themes")

	tf, err := loader.Load("dracula")
	if err != nil {
		t.Fatalf("Load user theme: %v", err)
	}

	if tf.Name != "dracula" {
		t.Errorf("name = %q, want dracula", tf.Name)
	}
}
