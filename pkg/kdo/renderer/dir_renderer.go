package renderer

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

// DirSpec -
type DirSpec struct {
	Dir          string
	FileRenderer func(filename string) func(writer io.Writer) error
}

// DirRender -
func DirRender(glob string, specs ...DirSpec) func(io.Writer) error {
	if glob == "" {
		glob = "*.y*ml"
	}
	return func(in io.Writer) error {
		for _, r := range specs {

			var filenames []string
			err := filepath.Walk(r.Dir, func(file string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					match, err := filepath.Match(glob, path.Base(file))
					if err != nil {
						return err
					}
					if match {
						filenames = append(filenames, file)
					}
				}
				return nil
			})

			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
			var processors []func(writer io.Writer) error
			for _, filename := range filenames {
				processors = append(processors, r.FileRenderer(filename))
			}
			for _, processor := range processors {
				writer := &YamlWriter{Writer: in}
				err := processor(writer)
				if err != nil {
					return err
				}
				writer.Write([]byte("\n"))
			}
		}
		return nil
	}
}
