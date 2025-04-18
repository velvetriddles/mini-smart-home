FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o device .

# Use distroless for the final container
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/device /app/

# Set the entrypoint
ENTRYPOINT ["/app/device"] 