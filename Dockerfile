FROM golang:1.23-rc-alpine AS builder

WORKDIR /go/src/sample-exchange

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the backend
COPY backend ./backend

# Patch the db.go file to use environment variables
RUN sed -i 's/dsn := "host=localhost user=postgres password=postgres dbname=sample_exchange port=5432 sslmode=disable"/dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"))/' ./backend/db/db.go || true
RUN sed -i 's/import (/import (\n\t"os"/' ./backend/db/db.go || true

# Build the backend
RUN CGO_ENABLED=0 GOOS=linux go build -o quixit ./backend

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /go/src/sample-exchange/quixit .

# Copy the pre-built frontend
COPY frontend/dist /app/frontend/dist

# Create necessary directories
RUN mkdir -p /app/uploads /app/storage && \
    chown -R nobody:nobody /app

USER nobody
EXPOSE 8080

# Run the binary
CMD ["./quixit"] 