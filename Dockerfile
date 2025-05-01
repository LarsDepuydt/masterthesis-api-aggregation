# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Define an argument for the application source directory relative to the context
ARG APP_SRC=.

WORKDIR /app

# Copy only necessary files first for layer caching
COPY ${APP_SRC}/go.mod ${APP_SRC}/go.sum ./
RUN go mod download

# Copy the specific application source code using the build argument
COPY ${APP_SRC} ./

# Build the application binary, ensure static linking
# The '.' at the end refers to the WORKDIR (/app) where the source was copied
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app-binary .

# Stage 2: Runner
FROM alpine:latest

# Create a non-root user for security
RUN adduser -D -u 10001 appuser
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app-binary .
# Change ownership if using a non-root user
RUN chown appuser:appuser /app/app-binary

# Switch to the non-root user
USER appuser

# Default command (can be overridden in docker-compose)
CMD ["./app-binary"]
