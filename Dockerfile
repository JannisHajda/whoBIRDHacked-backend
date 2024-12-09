FROM golang:1.23 AS builder

WORKDIR /app

# Copy all files in the directory except those excluded in .dockerignore
COPY . .

# Download dependencies and build the application
RUN go mod download
RUN go build -o server ./cmd/backend

FROM debian:bookworm-slim

# Install required system packages
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built server and all necessary files
COPY --from=builder /app/server .
COPY --from=builder /app .

EXPOSE 8080

CMD ["./server"]
