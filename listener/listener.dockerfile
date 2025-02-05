# Build a small image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY listener ./

# Command to run the executable
CMD ["./listener"]
