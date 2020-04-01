package renderer

import (
	"io/ioutil"
	"path"
	"path/filepath"
)

// Files -
type Files struct {
	Dir string
}

// Glob -
func (f Files) Glob(pattern string) map[string][]byte {
	result := make(map[string][]byte)
	matches, err := filepath.Glob(path.Join(f.Dir, pattern))
	if err != nil {
		return result
	}
	for _, match := range matches {
		data, err := ioutil.ReadFile(match)
		if err == nil {
			p, err := filepath.Rel(f.Dir, match)
			if err == nil {
				result[p] = data
			} else {
			}
		} else {
		}
	}
	return result
}

// Get -
func (f Files) Get(name string) string {
	data, err := ioutil.ReadFile(path.Join(f.Dir, name))
	if err != nil {
		return err.Error()
	}
	return string(data)
}
