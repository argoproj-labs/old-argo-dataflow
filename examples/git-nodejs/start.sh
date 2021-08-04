#!/bin/sh

echo "Installing dependencies"
mkdir -p ./cache
npm install --cache ./cache
echo "Running handler"
node index.js
