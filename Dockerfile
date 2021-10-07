#syntax=docker/dockerfile:1.2
# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
# shell scripts don't kindly send signals down to their sub-processes, but we can use dumb-init to
# achive this important support
RUN wget -O /tmp/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64

FROM golang:1.16 AS make
# python
RUN apt update
RUN apt install -y software-properties-common
RUN apt install -y python3
RUN apt install -y python3-pip
# python lint
RUN pip3 install autopep8
# kubectl: apt's version does not support --load_restrictor
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN cp kubectl /usr/local/bin
RUN chmod +x /usr/local/bin/kubectl
# squash warnings
RUN kubectl config set-context none
RUN kubectl config use-context none
# protoc
RUN apt install -y protobuf-compiler
# `make pre-commit` will install remaining tools
ADD . /root/go/src/github.com/argoproj-labs/argo-dataflow
WORKDIR /root/go/src/github.com/argoproj-labs/argo-dataflow
RUN make pre-commit

FROM builder AS controller-builder
ARG VERSION=unset
COPY api/ api/
COPY shared/ shared/
COPY manager/ manager/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/argoproj-labs/argo-dataflow/shared/util.version=${VERSION}'" -o bin/manager ./manager

FROM gcr.io/distroless/static:nonroot AS controller
WORKDIR /
COPY --from=controller-builder /workspace/bin/manager .
USER 9653:9653
ENTRYPOINT ["/manager"]

FROM builder AS runner-builder
ARG VERSION=unset
COPY kill/ kill/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/kill ./kill
COPY prestop/ prestop/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/prestop ./prestop
COPY api/ api/
COPY shared/ shared/
COPY sdks/golang sdks/golang
COPY runner/ runner/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 go build -ldflags="-s -w -X 'github.com/argoproj-labs/argo-dataflow/shared/util.version=${VERSION}'" -o bin/runner ./runner

FROM gcr.io/distroless/base:nonroot AS runner
WORKDIR /
COPY --from=runner-builder /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY runtimes runtimes
COPY --from=runner-builder /workspace/bin/kill /bin/kill
COPY --from=runner-builder /workspace/bin/prestop /bin/prestop
COPY --from=runner-builder /workspace/bin/runner .
USER 9653:9653
ENTRYPOINT ["/runner"]

FROM builder AS testapi-builder
COPY testapi/ testapi/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/testapi ./testapi

FROM gcr.io/distroless/base:nonroot AS testapi
WORKDIR /
COPY --from=testapi-builder /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=testapi-builder /workspace/bin/testapi .
USER 9653:9653
ENTRYPOINT ["/testapi"]

FROM golang:1.16-alpine AS golang1-16
COPY --from=builder /tmp/dumb-init /dumb-init
RUN chmod +x /dumb-init
RUN mkdir /.cache
ENV GO111MODULE=off
ADD sdks/golang /go/src/github.com/argoproj-labs/argo-dataflow/sdks/golang
ADD runtimes/golang1-16 /workspace
RUN chown -R 9653 /.cache /workspace
WORKDIR /workspace
USER 9653:9653
RUN go build ./...
ENTRYPOINT ["/dumb-init", "--"]
CMD ["/workspace/entrypoint.sh"]

FROM openjdk:16 AS java16
COPY --from=builder /tmp/dumb-init /dumb-init
RUN chmod +x /dumb-init
ADD runtimes/java16 /workspace
RUN chown -R 9653 /workspace
WORKDIR /workspace
USER 9653:9653
RUN javac *.java
ENTRYPOINT ["/dumb-init", "--"]
CMD ["/workspace/entrypoint.sh"]

FROM python:3.9-alpine AS python3-9
COPY --from=builder /tmp/dumb-init /dumb-init
RUN chmod +x /dumb-init
RUN mkdir /.cache /.local
ADD runtimes/python3-9 /workspace
ADD dsls/python /workspace/.dsl
RUN cd /workspace/.dsl && python3 -m pip install --use-feature=in-tree-build .
ADD sdks/python /workspace/.sdk
RUN apk add --no-cache gcc musl-dev
RUN cd /workspace/.sdk && python3 -m pip install --use-feature=in-tree-build .
RUN apk del --purge gcc musl-dev
RUN chown -R 9653 /.cache /.local /workspace
WORKDIR /workspace
USER 9653:9653
ENV PYTHONUNBUFFERED 1
ENTRYPOINT ["/dumb-init", "--"]
CMD ["/workspace/entrypoint.sh"]

FROM node:16-alpine AS node16
COPY --from=builder /tmp/dumb-init /dumb-init
RUN chmod +x /dumb-init
RUN mkdir /.cache /.local
ADD runtimes/node16 /workspace
RUN chown -R 9653 /.cache /.local /workspace
WORKDIR /workspace
USER 9653:9653
RUN npm install --cache /.cache
ENTRYPOINT ["/dumb-init", "--"]
CMD ["/workspace/entrypoint.sh"]