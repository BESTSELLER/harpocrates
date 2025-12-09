FROM golang:1.25.5-alpine@sha256:26111811bc967321e7b6f852e914d14bede324cd1accb7f81811929a6a57fea9 AS builder
WORKDIR $GOPATH/src/harpocrates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o /tmp/harpocrates

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412
RUN apk add --no-cache bash
COPY --from=builder /tmp/harpocrates /harpocrates
COPY docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/harpocrates"]
