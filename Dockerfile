FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-s -w" -o /bin/kontainer-engine-driver-rackspacespot .

FROM scratch
COPY --from=builder /bin/kontainer-engine-driver-rackspacespot /usr/bin/kontainer-engine-driver-rackspacespot
ENTRYPOINT ["/usr/bin/kontainer-engine-driver-rackspacespot"]
