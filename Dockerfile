# Multi-stage build for the dops MCP server
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /dops .

FROM alpine:3.20

RUN apk add --no-cache git bash

COPY --from=builder /dops /usr/local/bin/dops

# Default: run MCP server on stdio
ENTRYPOINT ["dops", "mcp", "serve"]
CMD ["--transport", "stdio"]
