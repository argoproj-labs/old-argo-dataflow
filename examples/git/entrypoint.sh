#!/bin/sh
set -eux

export GOCACHE=/go/.cache

go env
go run .