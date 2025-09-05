FROM --platform=$BUILDPLATFORM golang:1.24.6 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /usr/src/arcus-cli
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd/ cmd/
COPY config/ config/
COPY internal/ internal/
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/arcus-cli

FROM gcr.io/distroless/static-debian12
COPY --from=builder /go/bin/arcus-cli /
USER nonroot:nonroot
ENTRYPOINT ["/arcus-cli"]
