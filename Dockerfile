FROM --platform=$BUILDPLATFORM golang:1.24.6 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /usr/src/arcusctl
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd/ cmd/
COPY config/ config/
COPY internal/ internal/
COPY main.go main.go
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/arcusctl

FROM gcr.io/distroless/static-debian12
COPY --from=builder /go/bin/arcusctl /
USER nonroot:nonroot
ENTRYPOINT ["/arcusctl"]
