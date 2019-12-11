FROM golang:1.13.1 as build
WORKDIR $GOPATH/src/bitbucket.org/bestsellerit/harpocrates
COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go mod vendor
RUN GO111MODULE=on CGO_ENABLED=0 go install -mod=vendor

FROM alpine
WORKDIR /
COPY --from=build /go/bin/harpocrates /

COPY docker-entrypoint.sh /
RUN chmod +x "/docker-entrypoint.sh"
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/harpocrates"]