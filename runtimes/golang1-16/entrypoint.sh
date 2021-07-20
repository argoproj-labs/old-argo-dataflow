#!/bin/sh
set -eux

trap 'echo terminate' 15

pwd

cp /var/run/argo-dataflow/handler handler.go

go env
go run .