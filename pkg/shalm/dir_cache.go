package shalm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
)

type cacheMetaData struct {
	ETag       string `yaml:"etag,omitempty"`
	Generation uint32 `yaml:"generation,omitempty"`
	Name       string `yaml:"name,omitempty"`
}

// DirCache -
type DirCache struct {
	baseDir string
}

// NewDirCache -
func NewDirCache(baseDir string) *DirCache {
	return &DirCache{baseDir: baseDir}
}

// OpenDirCache -
type OpenDirCache = func(name string) (cachedDir string, err error)

// LoadDir -
type LoadDir = func(name string, targetDir func() (string, error), etagOld string) (etagNew string, err error)

// WrapDir -
func (d *DirCache) WrapDir(load LoadDir) OpenDirCache {
	return func(name string) (cachedDir string, err error) {
		md5Sum := md5.Sum([]byte(name))
		cacheDir := path.Join(d.baseDir, hex.EncodeToString(md5Sum[:]))
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return "", err
		}
		metaDataFile := path.Join(cacheDir, "metadata.yml")
		var metaData cacheMetaData

		if _, err := os.Stat(metaDataFile); err == nil {
			err := readYamlFile(metaDataFile, &metaData)
			if err != nil {
				return "", err
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
		contentDir := func(generation uint32) string {
			return path.Join(cacheDir, fmt.Sprintf("%x", generation))
		}
		oldDir := contentDir(metaData.Generation)
		newDir := ""
		targetDir := func() (string, error) {
			if len(newDir) != 0 {
				return newDir, nil
			}
			metaData.Generation++
			newDir = contentDir(metaData.Generation)
			if err := os.MkdirAll(newDir, 0755); err != nil {
				newDir = ""
				return "", err
			}
			return newDir, nil
		}

		etag, err := load(name, targetDir, metaData.ETag)
		if err != nil {
			return "", err
		}
		if etag == metaData.ETag {
			return contentDir(metaData.Generation), nil
		}
		if len(newDir) == 0 {
			return "", fmt.Errorf("No content written for %s", name)
		}
		metaData.Name = name
		metaData.ETag = etag
		if err := writeYamlFile(metaDataFile, metaData); err != nil {
			return "", err
		}
		_ = os.RemoveAll(oldDir)
		return contentDir(metaData.Generation), nil
	}
}

// OpenReaderCache -
type OpenReaderCache = func(name string) (io.ReadCloser, error)

// LoadReader -
type LoadReader = func(name string, etagOld string) (reader io.ReadCloser, etagNew string, err error)

// WrapReader -
func (d *DirCache) WrapReader(load LoadReader) OpenReaderCache {
	loadDir := func(name string, targetDir func() (string, error), etagOld string) (etagNew string, err error) {
		reader, etagNew, err := load(name, etagOld)
		if err != nil {
			return "", err
		}
		if etagNew == etagOld {
			return etagNew, nil
		}
		defer reader.Close()
		dir, err := targetDir()
		if err != nil {
			return "", err
		}
		writer, err := os.Create(path.Join(dir, "download"))
		if err != nil {
			return "", err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		return etagNew, err

	}
	wrapped := d.WrapDir(loadDir)
	return func(name string) (io.ReadCloser, error) {
		dir, err := wrapped(name)
		if err != nil {
			return nil, err
		}
		return os.Open(path.Join(dir, "download"))
	}

}
