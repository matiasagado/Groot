# Start from the official Go image for building the application
FROM golang:1.23 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY src/go.mod src/go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY src .

# Build the Go application
RUN go build -o server .

# Final stage for the production container
FROM debian:bullseye-slim

# Set the working directory
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/server .

# Expose the application's port
EXPOSE 1323

# Command to run the application
CMD ["./server"]
