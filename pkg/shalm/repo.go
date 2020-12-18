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
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"

	"github.com/k14s/starlark-go/starlark"
	"github.com/pkg/errors"
	shalmv1a2 "github.com/wonderix/shalm/api/v1alpha2"
	"github.com/wonderix/shalm/pkg/shalm/renderer"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// RepoListOptions -
type RepoListOptions struct {
	allNamespaces bool
	namespace     string
	genus         string
}

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, url string, options ...ChartOption) (ChartValue, error)
	// GetFromSpec -
	GetFromSpec(thread *starlark.Thread, spec *shalmv1a2.ChartSpec, options ...ChartOption) (ChartValue, error)
	// List -
	List(thread *starlark.Thread, k8s K8s, listOptions *RepoListOptions) ([]ChartValue, error)
}

type repoImpl struct {
	cacheDir string
	cache    OpenDirCache
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
	cache := dirCache.WrapDir(loadArchive)
	cache = openLocal(cache)
	cache = openURL(dirCache, httpClient, configs.Credentials, cache)
	cache = openHelm(dirCache, httpClient, configs.Credentials, cache)
	cache = openWithFragment(cache)
	cache = openWithCatalogs(configs.Catalogs, cache)
	r := &repoImpl{
		cacheDir: path.Join(homedir, ".shalm", "cache"),
		cache:    cache,
	}
	return r, nil
}

func openURL(dirCache *DirCache, client *http.Client, credentials []credential, dfltCache OpenDirCache) OpenDirCache {
	urlCache := dirCache.WrapDir(loadURL(client, credentials, extractArchive))
	return func(url string) (string, error) {
		if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
			return urlCache(url)
		}
		return dfltCache(url)
	}
}

func openHelm(dirCache *DirCache, client *http.Client, credentials []credential, cache OpenDirCache) OpenDirCache {
	helmCache := dirCache.WrapDir(loadURL(client, credentials, func(body io.Reader, dir string) error {
		out, err := os.Create(path.Join(dir, "index.yaml"))
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, body)
		return err
	}))
	return func(uri string) (string, error) {
		u, err := url.Parse(uri)
		if err != nil {
			return cache(uri)
		}
		if u.Scheme == "helm" {
			chart := path.Base(u.Path)
			u.Path = path.Join(path.Dir(u.Path), "index.yaml")
			u.Scheme = "https"
			dir, err := helmCache(u.String())
			if err != nil {
				return "", err
			}
			indexIn, err := os.Open(path.Join(dir, "index.yaml"))
			if err != nil {
				return "", err
			}
			defer indexIn.Close()
			var index renderer.Index
			decoder := yaml.NewDecoder(indexIn)
			err = decoder.Decode(&index)
			if err != nil {
				return "", err
			}
			entries, ok := index.Entries[chart]
			if !ok {
				return "", fmt.Errorf("chart %s not found in index", chart)
			}
			for _, entry := range entries {
				if !entry.Deprecated && len(entry.URLs) > 0 {
					return cache(entry.URLs[0])
				}
			}
		}
		return cache(uri)
	}
}

func openWithFragment(cache OpenDirCache) OpenDirCache {
	return func(uri string) (string, error) {
		u, err := url.Parse(uri)
		if err != nil {
			return cache(uri)
		}
		if u.Fragment != "" {
			fragment := u.Fragment
			u.Fragment = ""
			dir, err := cache(u.String())
			return dir + "/" + fragment, err
		}
		return cache(uri)
	}
}

func openWithCatalogs(catalogs []string, cache OpenDirCache) OpenDirCache {
	return func(uri string) (string, error) {
		if match := catalogURL.FindStringSubmatch(uri); match != nil {
			err := errors.New("No catalogs found")
			dir := ""
			for _, catalog := range catalogs {
				dir, err = cache(catalog + "/" + match[1])
				if err == nil {
					return dir, nil
				}
			}
			return "", err
		}
		return cache(uri)
	}
}

// Get -
func (r *repoImpl) Get(thread *starlark.Thread, url string, opts ...ChartOption) (ChartValue, error) {

	dir, err := r.cache(url)
	if err != nil {
		return nil, fmt.Errorf("Chart not found for url %s: %s", url, err.Error())
	}
	return newChart(thread, r, dir, append(opts, NewGenusAndVersion(url).AsOptions()...)...)
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
	options = append(options, WithNamespace(spec.Namespace), WithSuffix(spec.Suffix), WithArgs(ToStarlark(spec.Args).(starlark.Tuple)), WithValues(values), WithValues(kwargs))
	if spec.ChartURL != "" {
		chart, err := r.Get(thread, spec.ChartURL, options...)
		if err != nil {
			c = chart.(*chartImpl)
		}

	} else {
		c, err = newChartFromReader(thread, r, r.cacheDirForChart(spec.ChartTgz), bytes.NewReader(spec.ChartTgz), options...)
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newChartFromConfigMap(thread *starlark.Thread, r *repoImpl, configMap Object) (ChartValue, error) {
	dataJSON, ok := configMap.Additional["data"]
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
	version, err := newVersion(configMap.MetaData.Labels["shalm.wonderix.github.com/version"])
	if err != nil {
		return nil, err
	}
	gv := &GenusAndVersion{version: version, genus: configMap.MetaData.Labels["shalm.wonderix.github.com/genus"]}
	return newChartFromReader(thread, r, r.cacheDirForChart(tgz), bytes.NewReader(tgz), gv.AsOptions()...)
}

func (r *repoImpl) List(thread *starlark.Thread, k8s K8s, repoListOptions *RepoListOptions) ([]ChartValue, error) {
	requirement, err := labels.NewRequirement("shalm.wonderix.github.com/chart", selection.Equals, []string{"true"})
	if err != nil {
		return nil, err
	}
	listOptions := &ListOptions{
		LabelSelector: labels.NewSelector().Add(*requirement),
		AllNamespaces: repoListOptions.allNamespaces,
	}
	if len(repoListOptions.genus) != 0 {
		requirement, err := labels.NewRequirement("shalm.wonderix.github.com/genus", selection.Equals, []string{repoListOptions.genus})
		if err != nil {
			return nil, err
		}
		listOptions.LabelSelector = listOptions.LabelSelector.Add(*requirement)
	}
	k8sOptions := &K8sOptions{Quiet: true, Namespace: repoListOptions.namespace, ClusterScoped: repoListOptions.allNamespaces}
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
	charts := make([]ChartValue, 0)
	for _, o := range items {
		chart, err := newChartFromConfigMap(thread, r, o)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}
	return charts, nil
}

func newChartFromReader(thread *starlark.Thread, repo Repo, dir string, reader io.Reader, opts ...ChartOption) (*chartImpl, error) {
	if err := extractArchive(reader, dir); err != nil {
		return nil, err
	}
	return newChart(thread, repo, dir, opts...)
}

var invalidLabel = regexp.MustCompile("[^-A-Za-z0-9_.]")

func extractGenusAndVersion(name, version string) *GenusAndVersion {
	result := &GenusAndVersion{}
	vers, err := newVersion(version)
	if err == nil {
		result.version = vers
	}
	result.genus = invalidLabel.ReplaceAllString(name, "_")
	return result
}

var githubRelease = regexp.MustCompile("https://(github[^/]*/[^/]*/[^/]*)/releases/download/([^/]*)/([^/-]*)")
var githubArchive = regexp.MustCompile("https://(github[^/]*/[^/]*/[^/]*)/archive/(.*).zip")
var githubEnterpriseArchive = regexp.MustCompile("https://(github[^/]*)/api/v3/repos/([^/]*/[^/]*)/zipball/(.*)")
var otherURL = regexp.MustCompile("(https|http)://(.*)/(v{0,1}\\d+\\.\\d+\\.\\d+)")
var catalogURL = regexp.MustCompile("catalog:(.*)")

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
	if err := extractArchive(in, dir); err != nil {
		return "", err
	}
	return tag, nil
}

func loadURL(client *http.Client, credentials []credential, extract func(body io.Reader, dir string) error) func(name string, targetDir func() (string, error), etagOld string) (string, error) {
	return func(url string, targetDir func() (string, error), etagOld string) (string, error) {
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
		dir, err := targetDir()
		if err != nil {
			return "", err
		}
		err = extract(res.Body, dir)
		if err != nil {
			return "", err
		}
		etag := res.Header.Get("Etag")
		if len(etag) == 0 {
			etag = fmt.Sprintf("%x", time.Now().Unix())
		}
		return etag, nil
	}
}

func openLocal(openArchive OpenDirCache) OpenDirCache {
	return func(url string) (cachedDir string, err error) {
		if stat, err := os.Stat(url); err == nil {
			if stat.IsDir() {
				return url, nil
			}
		}
		return openArchive(url)
	}
}

func extractArchive(in io.Reader, dir string) error {
	prefix := chartDirExpr
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

// AddFlags -
func (s *RepoListOptions) AddFlags(flagsSet *pflag.FlagSet) {
	flagsSet.BoolVarP(&s.allNamespaces, "all-namespaces", "A", false, "List charts in all namespaces")
	flagsSet.StringVarP(&s.namespace, "namespace", "n", "default", "namespace")
	flagsSet.StringVarP(&s.genus, "genus", "g", "", "Search for package with the given genus")

}
