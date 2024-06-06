# Start with the official Golang image
FROM golang:1.20-alpine

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

# Build the Go app
RUN go build -o main main.go

# Ensure the main binary is executable
RUN chmod +x main

# Expose the port that your app will run on
EXPOSE 5000

# Command to run the executable
CMD ["./main"]
