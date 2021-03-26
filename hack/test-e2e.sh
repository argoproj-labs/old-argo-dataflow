#!/bin/bash
set -eu -o pipefail

cd examples

ls *.yaml | while read f; do
  kubectl delete pipeline --all > /dev/null
  echo " ▶ TEST $f"
  kubectl apply -f "$f"
  kubectl wait pipeline --all --for condition=Available
done