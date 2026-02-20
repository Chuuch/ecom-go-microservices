FROM golang:1.25.7-alpine3.23 AS builder

# set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# copy the go mod and sum files
WORKDIR /build

# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o service cmd/main.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/service .

# Step 2 build a small image
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from /dist into our base image
COPY --from=builder /dist/service .

# Copy config directory from builder
COPY --from=builder /build/config ./config

# Expose port 5000
EXPOSE 5000

# Command to run the executable
ENTRYPOINT [ "./service" ]