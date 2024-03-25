# Use an official Golang runtime as the base image
FROM golang:1.21.7

COPY . /app
WORKDIR /app

# Copy the .env file
COPY .env /app/.env

# Load environment variables from the .env file
ENV $(cat /app/.env | xargs)

RUN go build -o main .

CMD ["./main"]
