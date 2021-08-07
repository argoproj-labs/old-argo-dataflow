#!/bin/sh
set -eux

cp /var/run/argo-dataflow/handler handler.js

node index.js
