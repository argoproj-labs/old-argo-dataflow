#!/bin/sh

echo "Installing dependencies in virtual environment"
python3 -m venv venv
$(pwd)/venv/bin/pip3 --no-cache-dir install -r requirements.txt
echo "Running handler"
$(pwd)/venv/bin/python3 main.py