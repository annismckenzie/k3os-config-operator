# syntax=docker/dockerfile:1-experimental

FROM --platform=${BUILDPLATFORM} golang:1.16.3-alpine AS base

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
COPY cmd/update_k3osnodes_secret/main.go cmd/update_k3osnodes_secret/main.go

# Build
ENV GO111MODULE=on
RUN --mount=type=cache,target=/root/.cache/go-build \
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=readonly -a -o update_k3osnodes_secret cmd/update_k3osnodes_secret/main.go

# use distroless as minimal base image to package the binary
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/update_k3osnodes_secret .
USER nonroot:nonroot

ENTRYPOINT ["/update_k3osnodes_secret"]
