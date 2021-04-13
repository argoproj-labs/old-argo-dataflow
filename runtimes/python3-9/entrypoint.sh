#!/bin/sh
set -eux

cp /var/run/argo-dataflow/handler handler.py

pip3 install -r requirements.txt
python3 main.py