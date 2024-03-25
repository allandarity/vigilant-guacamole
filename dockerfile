# Use an official Golang runtime as the base image
FROM golang:1.21.7

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o main .

# Set environment variable "host"
ENV host="http://192.168.1.157:8096"

# Command to run the executable
CMD ["./main"]
