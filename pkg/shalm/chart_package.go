package shalm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

func (c *chartImpl) Package(writer io.Writer, helmFormat bool) error {
	if helmFormat {
		return c.packageHelm(writer)
	}
	return c.packageTgz(writer)
}

func (c *chartImpl) packageHelm(writer io.Writer) error {
	gz := gzip.NewWriter(writer)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	if err := writeFile(tw, path.Join(c.clazz.Name, "Chart.yaml"), func(w io.Writer) error {
		encode := yaml.NewEncoder(w)
		return encode.Encode(c.clazz)
	}); err != nil {
		return err
	}
	var args []string
	if c.initFunc != nil {
		args = make([]string, c.initFunc.NumParams()-1)
		for i := 1; i < c.initFunc.NumParams(); i++ {
			args[i-1], _ = c.initFunc.Param(i)
		}
	}

	if err := writeFile(tw, path.Join(c.clazz.Name, "templates", "chart.yaml"), func(w io.Writer) error {
		buf := &bytes.Buffer{}
		b := base64.NewEncoder(base64.StdEncoding, buf)
		err := c.packageTgz(b)
		if err != nil {
			return err
		}
		b.Close()
		t := template.Must(template.New("chart").Parse(chartTemplate))
		return t.Execute(w, map[string]interface{}{
			"tag":      DockerTag(),
			"chartTgz": buf.String(),
			"name":     c.clazz.Name,
			"args":     args,
			"version":  strings.ReplaceAll(c.clazz.Version, ".", "-"),
		})
	}); err != nil {
		return err
	}
	if err := writeFile(tw, path.Join(c.clazz.Name, "values.yaml"), func(w io.Writer) error {
		t := template.Must(template.New("chart").Parse(valuesTemplate))
		return t.Execute(w, map[string]interface{}{
			"args": args,
		})
	}); err != nil {
		return err
	}

	return nil
}

func writeFile(tw *tar.Writer, name string, f func(io.Writer) error) error {
	buf := &bytes.Buffer{}
	if err := f(buf); err != nil {
		return err
	}
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(buf.Len()),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (c *chartImpl) packageTgz(writer io.Writer) error {
	gz := gzip.NewWriter(writer)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	return c.walk(func(file string, size int64, body io.Reader, err error) error {
		hdr := &tar.Header{
			Name: path.Join(c.clazz.Name, file),
			Mode: 0644,
			Size: size,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(tw, body); err != nil {
			return err
		}
		return nil
	})
}

var chartDirExpr = regexp.MustCompile("^[^/]*/")

func tarExtract(in io.Reader, dir string, prefix *regexp.Regexp) error {
	tr := tar.NewReader(in)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		if !prefix.MatchString(hdr.Name) {
			continue
		}
		fn := path.Join(dir, prefix.ReplaceAllString(hdr.Name, ""))
		if err := os.MkdirAll(path.Dir(fn), 0755); err != nil {
			return err
		}
		out, err := os.Create(fn)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			log.Fatal(err)
		}
		out.Close()
	}
	return nil
}

const valuesTemplate = `
{{ range $key, $value := .args }}
{{ $value }}: ~
{{ end }}
`
const chartTemplate = `---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .name }}-{{ .version }}-apply
  annotations:
    "helm.sh/hook": "post-install,post-upgrade"
spec:
  template:
    spec:
      containers:
      - name: {{ .name }}-{{ .version }}-apply
        image: wonderix/shalm:{{ .tag }}
        command: ["/app/shalm"]
        args: 
        - apply
        - '/tmp/charts/chart.tgz'
        {{ range $key, $value := .args -}}
        {{ printf "{{ if .Values.%s -}}" $value }}
        - '--set'
        - '{{ $value }}={{ printf "{{ .Values.%s }}" $value }}'
        {{ "{{- end }}" }}
        {{- end }}
        volumeMounts:
        - name: chart-volume
          mountPath: /tmp/charts
        env:
        - name: SHALM_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      restartPolicy: Never
      volumes:
      - name: chart-volume
        configMap:
          name: {{ .name }}-{{ .version }}
  backoffLimit: 4
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .name }}-{{ .version }}-delete
  annotations:
    "helm.sh/hook": "pre-delete"
spec:
  template:
    spec:
      containers:
      - name: {{ .name }}-{{ .version }}-delete
        image: wonderix/shalm:{{ .tag }}
        command: ["/app/shalm"]
        args: 
        - delete
        - '/tmp/charts/chart.tgz'
        {{ range $key, $value := .args -}}
        {{ printf "{{ if .Values.%s -}}" $value }}
        - '--set'
        - '{{ $value }}={{ printf "{{ .Values.%s }}" $value }}'
        {{ "{{- end }}" }}
        {{- end }}
        volumeMounts:
        - name: chart-volume
          mountPath: /tmp/charts
        env:
        - name: SHALM_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      restartPolicy: Never
      volumes:
      - name: chart-volume
        configMap:
          name: {{ .name }}-{{ .version }}
  backoffLimit: 4
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .name }}-{{ .version }}
binaryData:
  chart.tgz: {{ .chartTgz }}
`
