# Build a small image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install client
RUN apk add --no-cache curl postgresql-client  # <-- Add postgresql-client here

# Install migration tool
RUN apk add --no-cache curl \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate

# Copy the pre-built binary file from the previous build
COPY authentication ./

# Copy migration files and entrypoint
COPY migrations ./migrations
COPY entrypoint.sh .

# Set permissions and entrypoint
RUN chmod +x entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]

# Expose port 80
EXPOSE 80

# Command to run the executable
CMD ["./authentication"]