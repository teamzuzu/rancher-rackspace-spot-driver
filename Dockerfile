FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod tidy

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-s -w" -o /bin/rancher-rackspace-spot-driver .

FROM scratch
COPY --from=builder /bin/rancher-rackspace-spot-driver /usr/bin/rancher-rackspace-spot-driver
ENTRYPOINT ["/usr/bin/rancher-rackspace-spot-driver"]
