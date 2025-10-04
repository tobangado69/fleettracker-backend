#!/bin/bash
# Seed database with test data
# Run this from the backend directory

echo "🌱 Seeding database with test data..."

# Set DATABASE_URL to connect to postgres inside Docker network
export DATABASE_URL="postgres://fleettracker:password123@host.docker.internal:5432/fleettracker?sslmode=disable"

# Run the seed command
go run cmd/seed/main.go

if [ $? -eq 0 ]; then
    echo "✅ Database seeded successfully!"
else
    echo "❌ Failed to seed database"
    echo "Try: docker exec -it fleettracker-postgres psql -U fleettracker -d fleettracker"
    echo "Then manually run seed SQL commands"
fi

