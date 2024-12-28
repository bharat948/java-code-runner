# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache build-base

# Set working directory
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application with minimal size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Final stage
FROM alpine:latest

# Install OpenJDK from Alpine package and set JAVA_HOME
RUN apk add --no-cache openjdk11 --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community && \
    echo "export JAVA_HOME=/usr/lib/jvm/java-11-openjdk" >> /etc/profile && \
    echo "export PATH=$PATH:$JAVA_HOME/bin" >> /etc/profile

# Set environment variables
ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk
ENV PATH="${PATH}:${JAVA_HOME}/bin"

# Copy the compiled Go binary from builder stage
COPY --from=builder /app/main /app/main

# Set working directory
WORKDIR /app

# Create a directory for temporary files with appropriate permissions
RUN mkdir -p /app/temp && chmod 777 /app/temp

# Expose service port
EXPOSE 8080

# Run the application
CMD ["./main"]