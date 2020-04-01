package test

import (
	"io/ioutil"
	"os"
	"path"
)

// TestDir -
type TestDir string

// NewTestDir -
func NewTestDir() TestDir {
	dir, err := ioutil.TempDir("", "shalm")
	if err != nil {
		panic(err)
	}
	return TestDir(dir)
}

// Remove -
func (t TestDir) Remove() error {
	return os.RemoveAll(string(t))
}

// Join -
func (t TestDir) Join(parts ...string) string {
	parts = append([]string{t.Root()}, parts...)
	return path.Join(parts...)
}

// Root -
func (t TestDir) Root() string {
	return string(t)
}

// MkdirAll -
func (t TestDir) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(t.Join(path), mode)
}

// WriteFile -
func (t TestDir) WriteFile(path string, content []byte, mode os.FileMode) error {
	return ioutil.WriteFile(t.Join(path), content, mode)
}
