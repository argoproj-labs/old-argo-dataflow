#!/bin/sh
set -eux

export PIP_DOWNLOAD_CACHE=/tmp/pip/cache

pip3 install -r requirements.txt
python3 main.py