# Use the official Golang image as the base image
FROM golang:1.20-alpine AS builder

# Set environment variables for JWT and MySQL (used in the build stage)
ENV JWT_SECRET="your_jwt_secret"
ENV MYSQL_USER="root"
ENV MYSQL_PASSWORD="root"

# Install MySQL client and necessary tools
RUN apk add --no-cache bash mysql-client

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files to cache dependencies
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN go build -o shareit ./main.go

# Build the background worker
RUN go build -o worker ./background_worker/worker.go

# Use a lightweight Alpine image for the final container
FROM alpine:latest

# Install MySQL client and necessary tools in the final image
RUN apk add --no-cache bash mysql-client mysql mysql-server

# Set the working directory in the final image
WORKDIR /app

# Copy the built binaries from the builder stage
COPY --from=builder /app/shareit /app/worker /app/

# Copy the init.sql file to set up the database schema
COPY init.sql /docker-entrypoint-initdb.d/

# Set environment variables for the final container
ENV DB_USER="root"
ENV DB_PASSWORD="root"
ENV DB_HOST="localhost"
ENV DB_NAME="shareit"
ENV REDIS_URL="rediss://default:AYGsAAIjcDE2MmFmN2MwM2M2NzA0YWMzYjhhODM2N2MzMzgxMjhiN3AxMA@complete-feline-33196.upstash.io:6379"

# Expose the port that the Go app runs on
EXPOSE 8080

# Initialize the MySQL database
RUN mkdir /var/lib/mysql && chown -R mysql:mysql /var/lib/mysql && mysql_install_db --user=mysql

# Start MySQL, then run the app and background worker
CMD ["sh", "-c", "mysqld --init-file=/docker-entrypoint-initdb.d/init.sql & sleep 5 && ./shareit & ./worker"]
