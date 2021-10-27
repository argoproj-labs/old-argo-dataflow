#!/bin/sh
set -eux

pwd

cp /var/run/argo-dataflow/handler handler.go

go env
go run .