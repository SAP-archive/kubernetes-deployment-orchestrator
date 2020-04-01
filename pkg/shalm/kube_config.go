package shalm

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
)

func kubeConfigFromContent(content string) (string, error) {
	c := []byte(content)
	md5Sum := md5.Sum(c)
	filename := path.Join(os.TempDir(), hex.EncodeToString(md5Sum[:])+".kubeconfig")
	err := ioutil.WriteFile(filename, c, 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}
