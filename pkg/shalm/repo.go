package shalm

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	netUrl "net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/k14s/starlark-go/starlark"
	"github.com/pkg/errors"
	shalmv1a2 "github.com/wonderix/shalm/api/v1alpha2"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// RepoListOptions -
type RepoListOptions struct {
	AllNamespaces bool
	Namespace     string
}

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, url string, options ...ChartOption) (ChartValue, error)
	// GetFromSpec -
	GetFromSpec(thread *starlark.Thread, spec *shalmv1a2.ChartSpec, options ...ChartOption) (ChartValue, error)
	// List -
	List(thread *starlark.Thread, k8s K8s, listOptions *RepoListOptions) ([]Chart, error)
}

type repoImpl struct {
	cacheDir    string
	openURL     OpenDirCache
	openArchive OpenDirCache
}

var _ Repo = &repoImpl{}

const (
	customMediaType = "application/tar"
)

// NewRepo -
func NewRepo(config ...RepoConfig) (Repo, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configs := &repoConfigs{}
	for _, cfg := range config {
		err = cfg(configs)
		if err != nil {
			return nil, err
		}
	}
	httpClient := &http.Client{
		Timeout: time.Second * 60,
	}
	dirCache := NewDirCache(path.Join(homedir, ".shalm", "cache-etag"))
	r := &repoImpl{
		cacheDir:    path.Join(homedir, ".shalm", "cache"),
		openURL:     dirCache.WrapDir(loadURL(httpClient, configs.Credentials)),
		openArchive: dirCache.WrapDir(loadArchive),
	}
	return r, nil
}

// Get -
func (r *repoImpl) Get(thread *starlark.Thread, url string, opts ...ChartOption) (ChartValue, error) {

	co := chartOptions(opts)

	proxyFunc := func(chart *chartImpl, err error) (ChartValue, error) {
		return chart, err
	}
	if co.proxyMode != ProxyModeOff {
		proxyFunc = func(chart *chartImpl, err error) (ChartValue, error) {
			if err != nil {
				return nil, err
			}
			return newChartProxy(chart, url, co.proxyMode, co.args, co.kwargs)
		}
	}

	if isValidShalmURL(url) {
		return proxyFunc(r.newChartFromURL(thread, url, opts...))
	}
	if stat, err := os.Stat(url); err == nil {
		if stat.IsDir() {
			return proxyFunc(newChart(thread, r, url, opts...))
		}
		dir, err := r.openArchive(url)
		if err != nil {
			return nil, err
		}
		return proxyFunc(newChart(thread, r, dir, opts...))
	}
	return nil, fmt.Errorf("Chart not found for url %s", url)
}

func (r *repoImpl) cacheDirForChart(data []byte) string {
	md5Sum := md5.Sum(data)
	cacheDir := path.Join(r.cacheDir, hex.EncodeToString(md5Sum[:]))
	os.RemoveAll(cacheDir)
	return cacheDir
}

func (r *repoImpl) GetFromSpec(thread *starlark.Thread, spec *shalmv1a2.ChartSpec, options ...ChartOption) (ChartValue, error) {
	var c *chartImpl
	var err error
	kwargs, err := spec.GetKwArgs()
	if err != nil {
		return nil, err
	}
	values, err := spec.GetValues()
	if err != nil {
		return nil, err
	}
	options = append(options, WithNamespace(spec.Namespace), WithSuffix(spec.Suffix), WithArgs(ToStarlark(spec.Args).(starlark.Tuple)), WithValues(values),
		WithKwArgs(kwargsToStarlark(kwargs)))
	if spec.ChartURL != "" {
		c, err = r.newChartFromURL(thread, spec.ChartURL, options...)

	} else {
		c, err = newChartFromReader(thread, r, r.cacheDirForChart(spec.ChartTgz), bytes.NewReader(spec.ChartTgz), chartDirExpr, options...)
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newChartFromConfigMap(thread *starlark.Thread, r *repoImpl, o Object) (Chart, error) {
	dataJSON, ok := o.Additional["data"]
	if !ok {
		return nil, fmt.Errorf("Invalid config map")
	}
	var data map[string]string
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return nil, err
	}
	tgz, err := base64.StdEncoding.DecodeString(data["chart"])
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return nil, err
	}
	return newChartFromReader(thread, r, r.cacheDirForChart(tgz), bytes.NewReader(tgz), chartDirExpr)
}

func (r *repoImpl) List(thread *starlark.Thread, k8s K8s, repoListOptions *RepoListOptions) ([]Chart, error) {
	requirement, err := labels.NewRequirement("shalm.wonderix.github.com/chart", selection.Equals, []string{"true"})
	if err != nil {
		return nil, err
	}
	listOptions := &ListOptions{
		LabelSelector: labels.NewSelector().Add(*requirement),
		AllNamespaces: repoListOptions.AllNamespaces,
	}
	k8sOptions := &K8sOptions{Quiet: true, Namespace: repoListOptions.Namespace, Namespaced: !repoListOptions.AllNamespaces}
	obj, err := k8s.List("configmaps", k8sOptions, listOptions)
	if err != nil {
		return nil, err
	}
	itemsJSON := obj.Additional["items"]
	var items []Object
	err = json.Unmarshal(itemsJSON, &items)
	if err != nil {
		return nil, err
	}
	charts := make([]Chart, 0)
	for _, o := range items {
		chart, err := newChartFromConfigMap(thread, r, o)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}
	return charts, nil
}

func newChartFromReader(thread *starlark.Thread, repo Repo, dir string, reader io.Reader, prefix *regexp.Regexp, opts ...ChartOption) (*chartImpl, error) {
	if err := extract(reader, dir, prefix); err != nil {
		return nil, err
	}
	return newChart(thread, repo, dir, opts...)
}

func (r *repoImpl) newChartFromURL(thread *starlark.Thread, url string, opts ...ChartOption) (*chartImpl, error) {
	dir, err := r.openURL(url)
	if err != nil {
		return nil, err
	}
	return newChart(thread, r, dir, guessIDAndVersion(url, opts)...)
}

var invalidLabel = regexp.MustCompile("[^-A-Za-z0-9_.]")

func extractIDAndVersion(opts []ChartOption, name, version string) []ChartOption {
	vers, err := semver.ParseTolerant(version)
	name = invalidLabel.ReplaceAllString(name, "_")
	if err == nil {
		return append(opts, WithID(name), WithVersion(vers))
	}
	return append(opts, WithID(name))

}

var githubRelease = regexp.MustCompile("https://(github[^/]*/[^/]*/[^/]*)/releases/download/([^/]*)/([^/-]*)")
var githubArchive = regexp.MustCompile("https://(github[^/]*/[^/]*/[^/]*)/archive/(.*).zip")
var githubEnterpriseArchive = regexp.MustCompile("https://(github[^/]*)/api/v3/repos/([^/]*/[^/]*)/zipball/(.*)")
var otherURL = regexp.MustCompile("(https|http)://(.*)/(v{0,1}\\d+\\.\\d+\\.\\d+)")

func guessIDAndVersion(url string, opts []ChartOption) []ChartOption {
	var match []string
	if match = githubRelease.FindStringSubmatch(url); match != nil {
		return extractIDAndVersion(opts, match[1], match[2])
	}
	if match = githubArchive.FindStringSubmatch(url); match != nil {
		return extractIDAndVersion(opts, match[1], match[2])
	}
	if match = githubEnterpriseArchive.FindStringSubmatch(url); match != nil {
		return extractIDAndVersion(opts, match[1]+"/"+match[2], match[3])
	}
	if match = otherURL.FindStringSubmatch(url); match != nil {
		return extractIDAndVersion(opts, match[2], match[3])
	}
	return opts
}
func loadArchive(name string, targetDir func() (string, error), etagOld string) (string, error) {
	stat, err := os.Stat(name)
	if err != nil {
		return "", err
	}
	tag := stat.ModTime().String()
	if etagOld == tag {
		return tag, nil
	}
	dir, err := targetDir()
	if err != nil {
		return "", err
	}
	in, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer in.Close()
	if err := extract(in, dir, chartDirExpr); err != nil {
		return "", err
	}
	return tag, nil
}

func loadURL(client *http.Client, credentials []credential) func(name string, targetDir func() (string, error), etagOld string) (string, error) {
	return func(url string, targetDir func() (string, error), etagOld string) (string, error) {
		u, err := netUrl.Parse(url)
		if err != nil {
			return "", fmt.Errorf("Error parsing url %s: %v", url, err)
		}
		request, err := http.NewRequest(http.MethodGet, url, nil)
		maxMatch := 0
		var bestMatch *credential
		for _, cred := range credentials {
			if strings.HasPrefix(url, cred.URL) && len(cred.URL) > maxMatch {
				maxMatch = len(cred.URL)
				bestMatch = &cred
			}
		}
		if bestMatch != nil {
			if len(bestMatch.Token) != 0 {
				request.Header.Add("Authorization", "token "+bestMatch.Token)
			} else if len(bestMatch.Username) != 0 {
				request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(bestMatch.Username+":"+bestMatch.Password)))
			}
		}

		request.Header.Add("If-None-Match", etagOld)
		res, err := client.Do(request)
		if err != nil {
			return "", fmt.Errorf("Error fetching %s: %v", url, err)
		}
		if res.StatusCode == 304 {
			return etagOld, nil
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("Error fetching %s: status=%d", url, res.StatusCode)
		}
		defer res.Body.Close()
		prefix := chartDirExpr
		if u.Fragment != "" {
			path := u.Fragment
			if strings.HasSuffix(path, "/") {
				path = path + "/"
			}
			prefix = regexp.MustCompile("^" + chartDirExpr.String() + regexp.QuoteMeta(path))
		}
		dir, err := targetDir()
		if err != nil {
			return "", err
		}
		extract(res.Body, dir, prefix)
		etag := res.Header.Get("Etag")
		if len(etag) == 0 {
			etag = fmt.Sprintf("%x", time.Now().Unix())
		}
		return etag, nil
	}
}

func extract(in io.Reader, dir string, prefix *regexp.Regexp) error {
	reader := bufio.NewReader(in)
	testBytes, err := reader.Peek(64)
	if err != nil {
		return err
	}
	in = reader
	contentType := http.DetectContentType(testBytes)
	switch contentType {
	case "application/zip":
		return zipExtract(in, dir, prefix)
	case "application/octet-stream":
		return tarExtract(in, dir, prefix)
	case "application/x-gzip":
		in, err = gzip.NewReader(in)
		if err != nil {
			return err
		}
		return tarExtract(in, dir, prefix)
	}
	return errors.Errorf("Unsupported shalm archive type: %s", contentType)
}

func zipExtract(in io.Reader, dir string, prefix *regexp.Regexp) error {
	buf := &bytes.Buffer{}
	size, err := buf.ReadFrom(in)
	if err != nil {
		return err
	}
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), size)
	if err != nil {
		return err
	}
	for _, f := range r.File {

		if f.FileInfo().IsDir() {
			continue
		}
		if !prefix.MatchString(f.Name) {
			continue
		}
		fn := path.Join(dir, prefix.ReplaceAllString(f.Name, ""))

		if !strings.HasPrefix(fn, path.Clean(dir)+string(os.PathSeparator)) {
			return errors.Errorf("%s: illegal file path", fn)
		}
		if err := os.MkdirAll(path.Dir(fn), 0755); err != nil {
			return err
		}
		out, err := os.Create(fn)
		if err != nil {
			return err
		}
		tr, err := f.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			log.Fatal(err)
		}
		tr.Close()
		out.Close()
	}
	return nil

}
