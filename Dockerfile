FROM golang:1.21-alpine

# Install C-compiler required for go-sqlite3
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy mod files
COPY backend/go.mod ./
# Generate go.sum inside container to avoid host dependency issues
RUN go mod tidy

# Copy source
COPY backend/ .

# Build with CGO enabled
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN go build -o bank-server main.go database.go

EXPOSE 8080
CMD ["./bank-server"]
