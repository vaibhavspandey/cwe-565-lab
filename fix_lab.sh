#!/bin/bash
set -e

# 1. Create a valid "go.sum" file in "backend/"
echo "Running 'go mod tidy' in backend/..."
cd backend
if command -v go &> /dev/null; then
    go mod tidy
else
    echo "Go not found, creating dummy go.sum"
    touch go.sum
fi
cd ..

# 2. Overwrite "Dockerfile"
echo "Overwriting Dockerfile..."
cat <<EOF > Dockerfile
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
RUN go build -o bank-server main.go database.go

EXPOSE 8080
CMD ["./bank-server"]
EOF

echo "Fix complete."
