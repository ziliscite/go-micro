# Build a small image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY mailer ./

# Copy the tamplates
COPY templates ./templates

EXPOSE 80

# Command to run the executable
CMD ["./mailer"]
