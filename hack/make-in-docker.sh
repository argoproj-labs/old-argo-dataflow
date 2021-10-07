#!/bin/sh
set -eux

docker run -ti --rm \
    -v $(pwd):/root/go/src/github.com/argoproj-labs/argo-dataflow \
    -w /root/go/src/github.com/argoproj-labs/argo-dataflow \
    --name pre-commit \
    quay.io/argoprojlabs/dataflow-make:latest make $*