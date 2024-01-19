# Use the official Golang image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod .
COPY go.sum .

# Download and install Go dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port the app runs on
EXPOSE 80

# Command to run the executable
CMD ["./main"]
