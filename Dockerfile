FROM golang:latest AS builder

# Set build directory
WORKDIR /build

# Copy Go dependency files
COPY src/go.* .

# Download Go modules
RUN go mod download

# Copy the rest of the source code
COPY src/* .

# Build app
RUN go build -v -o redpin



FROM alpine:latest

# Copy compiled binary from previous stage
COPY --from=builder /build/redpin .

# Run redpin
CMD ["./redpin"]
