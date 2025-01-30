# Build a small image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY logger ./

# Dont expose ports, because this service will only be called through broker
EXPOSE 80
#EXPOSE 5001
#EXPOSE 8080

# Command to run the executable
CMD ["./logger"]
