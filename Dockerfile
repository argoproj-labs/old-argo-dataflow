#syntax=docker/dockerfile:1.2
# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM builder AS controller-builder
COPY .git/ .git/
COPY api/ api/
COPY shared/ shared/
COPY manager/ manager/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -a -ldflags="-s -w -X 'github.com/argoproj-labs/argo-dataflow/shared/util.message=$(git log -n1 --oneline)'" -o bin/manager ./manager

FROM gcr.io/distroless/static:nonroot AS controller
WORKDIR /
COPY --from=controller-builder /workspace/bin/manager .
USER 9653:9653
ENTRYPOINT ["/manager"]

FROM builder AS runner-builder
COPY .git/ .git/
COPY api/ api/
COPY shared/ shared/
COPY runner/ runner/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -a -ldflags="-s -w -X 'github.com/argoproj-labs/argo-dataflow/shared/util.message=$(git log -n1 --oneline)'" -o bin/runner ./runner
COPY kill/ kill/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -a -ldflags="-s -w" -o bin/kill ./kill
COPY prestop/ prestop/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -a -ldflags="-s -w" -o bin/prestop ./prestop

FROM gcr.io/distroless/static:nonroot AS runner
WORKDIR /
COPY runtimes runtimes
COPY --from=runner-builder /workspace/bin/runner .
COPY --from=runner-builder /workspace/bin/kill /bin/kill
COPY --from=runner-builder /workspace/bin/prestop /bin/prestop
USER 9653:9653
ENTRYPOINT ["/runner"]

FROM golang:1.16-alpine AS go1-16
RUN mkdir /.cache
ADD runtimes/go1-16 /workspace
RUN chown -R 9653 /.cache /workspace
WORKDIR /workspace
USER 9653:9653
RUN go build ./...
ENTRYPOINT ./entrypoint.sh

FROM openjdk:16 AS java16
ADD runtimes/java16 /workspace
RUN chown -R 9653 /workspace
WORKDIR /workspace
USER 9653:9653
RUN javac *.java
ENTRYPOINT ./entrypoint.sh

FROM python:3.9-alpine AS python3-9
RUN mkdir /.cache /.local
ADD runtimes/python3-9 /workspace
RUN chown -R 9653 /.cache /.local /workspace
WORKDIR /workspace
USER 9653:9653
RUN pip3 install -r requirements.txt
ENTRYPOINT ./entrypoint.sh
