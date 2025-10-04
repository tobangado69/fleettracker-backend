#!/bin/bash
# Run tests with proper database connection
# Run this from the backend directory

echo "🧪 Running unit tests..."

# Set DATABASE_URL to connect to postgres inside Docker network
export DATABASE_URL="postgres://fleettracker:password123@host.docker.internal:5432/fleettracker?sslmode=disable"

# Run the tests
go test -v ./internal/auth/... -count=1

if [ $? -eq 0 ]; then
    echo "✅ Tests passed!"
else
    echo "❌ Tests failed"
fi

