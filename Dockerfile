# syntax=docker/dockerfile:1-experimental

FROM --platform=${BUILDPLATFORM} golang:1.18.2-alpine AS base

WORKDIR /workspace
ENV CGO_ENABLED=0

# Copy the Go Modules manifests
COPY go.* .
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

FROM base AS builder
ARG TARGETOS
ARG TARGETARCH

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY config/ config/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
ENV GO111MODULE=on
RUN --mount=type=cache,target=/root/.cache/go-build \
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
