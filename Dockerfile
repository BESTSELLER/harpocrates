FROM golang:1.26rc3-alpine@sha256:343c20fd6876bfb5ba9f46b0a452008b7dced3804e424ff7ada0ceadafad5c55 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-w -s" -o /tmp/harpocrates

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN apk add --no-cache bash
COPY --from=builder /tmp/harpocrates /harpocrates

ENTRYPOINT ["/harpocrates"]
