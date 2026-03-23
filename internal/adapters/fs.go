package adapters

import (
	"io/fs"
	"os"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	ReadDir(path string) ([]os.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	Stat(path string) (os.FileInfo, error)
}

type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (f *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (f *OSFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (f *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *OSFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
