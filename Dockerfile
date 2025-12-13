# syntax=docker/dockerfile:1.4

# BUILD stage
FROM golang:1.25-alpine AS builder
RUN apk update && apk add --no-cache ca-certificates 

WORKDIR /app

# This step is cached efficiently if only source code changes
COPY src/go.mod src/go.sum ./
RUN go mod download 

COPY src/. .

# critical for static linking
ENV CGO_ENABLED=0
RUN go generate ./... 
RUN go build -ldflags="-s -w" -o /app/server ./cmd/server 

# RUNTIME stage
FROM alpine:3.22

RUN adduser -D nonroot 

RUN mkdir -p /var/pocketeer && \
    chown -R nonroot:nonroot /var/pocketeer

USER nonroot
WORKDIR /home/nonroot/

COPY --from=builder /app/server .
CMD ["./server"]
