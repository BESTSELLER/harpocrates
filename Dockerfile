FROM golang:1.20.7-alpine AS builder
WORKDIR $GOPATH/src/harpocrates
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o /tmp/harpocrates

FROM alpine
RUN apk add --no-cache bash
COPY --from=builder /tmp/harpocrates /harpocrates
COPY docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/harpocrates"]