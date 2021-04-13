#!/bin/sh
set -eux

cp /var/run/argo-dataflow/handler Handler.java

javac *.java
java -cp . Main