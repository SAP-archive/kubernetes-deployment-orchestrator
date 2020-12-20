#!/usr/bin/env bash
set -eu
unset DOLLAR
go run github.com/maxbrunsfeld/counterfeiter/v6 -o fake_k8s.go . K8s
sed -e 's|k8s\.||g' -i "" fake_k8s.go
sed -e 's|^.*"github.*k8s".*$||g' -i "" fake_k8s.go

