#!/bin/sh

echo "Installing dependencies in virtual environment"
pip3 --no-cache-dir install -r requirements.txt
echo "Running handler"
python3 main.py