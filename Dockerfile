FROM golang:1.21-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w" -o /bin/rancher-rackspace-spot-driver .

FROM scratch
COPY --from=builder /bin/rancher-rackspace-spot-driver /usr/bin/rancher-rackspace-spot-driver
ENTRYPOINT ["/usr/bin/rancher-rackspace-spot-driver"]
