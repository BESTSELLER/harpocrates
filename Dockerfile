FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/harpocrates
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /tmp/harpocrates

FROM alpine
COPY --from=builder /tmp/harpocrates /harpocrates
COPY docker-entrypoint.sh /
RUN chmod +x "/docker-entrypoint.sh"
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/harpocrates"]