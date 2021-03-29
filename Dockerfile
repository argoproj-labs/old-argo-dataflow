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
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -a -o manager main.go

FROM gcr.io/distroless/static:nonroot AS controller
WORKDIR /
COPY --from=controller-builder /workspace/manager .
USER nonroot:nonroot
ENTRYPOINT ["/manager"]

FROM builder AS runner-builder
COPY runner/main.go main.go
COPY api/ api/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -a -o runner main.go

FROM gcr.io/distroless/static:nonroot AS runner
WORKDIR /
COPY --from=runner-builder /workspace/runner .
USER nonroot:nonroot
ENTRYPOINT ["/runner"]
