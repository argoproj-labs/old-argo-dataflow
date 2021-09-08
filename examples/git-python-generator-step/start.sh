#!/bin/sh

echo "Installing dependencies in virtual environment"
pip3 --no-cache-dir install -r requirements.txt
pip3 list
ls -la
echo "Running handler"
python3 main.py