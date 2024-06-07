# Start with the official Golang image
FROM golang:1.20-alpine as builder

# Install build dependencies
RUN apk update && apk add --no-cache gcc musl-dev

# Set the environment variable to enable CGO
ENV CGO_ENABLED=1

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code and config.yaml into the container
COPY . .

# Build the Go app with output named 'main'
RUN go build -o main .

# Create a lightweight final image
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Set the working directory in the final image
WORKDIR /root/

# Copy the binary and config.yaml from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .

# Expose the port that your app will run on
EXPOSE 4000

# Command to run the executable
CMD ["./main"]
