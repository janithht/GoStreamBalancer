# Use the official Golang image as a builder stage to compile the application
FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o load-balancer-go .

# Start a new stage from scratch for a lightweight final image
FROM alpine:latest

LABEL maintainer="Janith Hathnagoda"

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage and the config file
COPY --from=builder /app/load-balancer-go .
COPY config.yaml .

# Expose port 9000 to the outside world
EXPOSE 9000

# Command to run the executable
CMD ["./load-balancer-go"]