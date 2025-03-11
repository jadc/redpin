FROM golang:latest AS builder

# Set build directory
WORKDIR /build

# Copy Go dependency files
COPY src/go.* .

# Download and verify Go modules
RUN go mod download
RUN go mod verify

# Copy the rest of the source code
COPY src/ .

# Build app
RUN CGO_ENABLED=0 GOOS=linux go build -v -o redpin



FROM scratch

# Copy necessary files from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/redpin .

# Run redpin
CMD ["./redpin"]
