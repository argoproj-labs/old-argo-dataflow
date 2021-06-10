#!/bin/sh
set -eux

cp /var/run/argo-dataflow/handler handler.py

python3 main.py