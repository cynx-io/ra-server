# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the files
COPY . .

# Build your binary
ENV CGO_ENABLED=0
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o ra main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/ra .

# Copy config and env files if needed
COPY config.json .
COPY .env .

# Expose the port your app uses
EXPOSE 5002

# Run the binary
CMD ["./ra"]
