# Build the manager binary
FROM golang:1.15.7 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

FROM builder AS controller-builder
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM gcr.io/distroless/static:nonroot AS controller
WORKDIR /
COPY --from=controller-builder /workspace/manager .
USER nonroot:nonroot
ENTRYPOINT ["/manager"]

FROM builder AS sidecar-builder
COPY sidecar/main.go main.go
COPY api/ api/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o sidecar main.go

FROM gcr.io/distroless/static:nonroot AS sidecar
WORKDIR /
COPY --from=sidecar-builder /workspace/sidecar .
USER nonroot:nonroot
ENTRYPOINT ["/sidecar"]

FROM builder AS cat-builder
COPY cat/main.go main.go
COPY api/ api/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o cat main.go

FROM gcr.io/distroless/static:nonroot AS cat
WORKDIR /
COPY --from=cat-builder /workspace/cat .
USER nonroot:nonroot
ENTRYPOINT ["/cat"]
