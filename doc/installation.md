# Installation

## Prerequisite

* Install `kubectl` e.g. using `brew install kubernetes-cli`
* Install `ytt` e.g. from `https://github.com/k14s/ytt/releases`
* Install `kapp` e.g. from `https://github.com/k14s/kapp/releases`

## Installing on MacOS


```bash
brew tap sap/tap
brew install kdo
```

## Install binary

* Download `kdo` (e.g. for mac os)

```bash
curl -L https://github.com/sap/kubernetes-deployment-orchestrator/releases/latest/download/kdo-binary-darwin.tgz | tar xzvf -
```

## Build `kdo` from source

* Install `go` e.g. using `brew install go`
* Install `kdo`

```bash
go get github.com/sap/kubernetes-deployment-orchestrator
```
