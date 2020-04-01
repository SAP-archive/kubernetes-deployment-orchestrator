# Installation

## Prerequisite

* Install `kubectl` e.g. using `brew install kubernetes-cli`
* Install `ytt` e.g. from `https://github.com/k14s/ytt/releases`
* Install `kapp` e.g. from `https://github.com/k14s/kapp/releases`

## Install binary

* Download `shalm` (e.g. for mac os)

```bash
curl -L https://github.com/wonderix/shalm/releases/latest/download/shalm-binary-darwin.tgz | tar xzvf -
```

## Build `shalm` from source

* Install `go` e.g. using `brew install go`
* Install `shalm`

```bash
go get github.com/wonderix/shalm
```
