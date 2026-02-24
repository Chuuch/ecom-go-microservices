FROM golang:1.25.7-alpine3.23 AS builder

ENV CGO_ENABLED=0

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app cmd/main.go

WORKDIR /dist

RUN cp /build/app .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY . .
COPY --from=builder /dist/app .

CMD ["./app"]