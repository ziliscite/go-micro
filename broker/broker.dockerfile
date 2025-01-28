## Base Go Image
#FROM golang:1.23.4-alpine AS builder
#
## Set working directory
#WORKDIR /app
#
## Add source code
#COPY . /app
#
## Build the binary and add environment variable through CGO_ENABLED
#RUN CGO_ENABLED=0 go build -o broker ./cmd/api
#
#RUN chmod +x /app/broker

## Not using the builder again, because we buildin binary alr

# Build a small image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/broker .

# Expose port 80
EXPOSE 80

# Command to run the executable
CMD ["./broker"]
