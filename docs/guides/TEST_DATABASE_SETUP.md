# Test Database Setup

This document explains how to set up a local PostgreSQL database for running tests.

## Prerequisites

- PostgreSQL installed locally
- PostgreSQL service running

## Quick Setup

### Option 1: Using default postgres user (Recommended)

1. **Start PostgreSQL service**:
   ```bash
   # Windows (if installed as service)
   net start postgresql-x64-18
   
   # Or start manually
   pg_ctl start -D "C:\Program Files\PostgreSQL\18\data"
   ```

2. **Create test database** (optional):
   ```sql
   -- Connect to PostgreSQL as postgres user
   psql -U postgres
   
   -- Create database (optional, tests will use postgres database)
   CREATE DATABASE fleettracker;
   ```

3. **Run tests**:
   ```bash
   go test -v ./...
   ```

### Option 2: Using custom user

1. **Create user and database**:
   ```sql
   -- Connect to PostgreSQL as postgres user
   psql -U postgres
   
   -- Create user
   CREATE USER fleettracker WITH PASSWORD 'password123';
   
   -- Create database
   CREATE DATABASE fleettracker OWNER fleettracker;
   
   -- Grant privileges
   GRANT ALL PRIVILEGES ON DATABASE fleettracker TO fleettracker;
   ```

2. **Run tests**:
   ```bash
   go test -v ./...
   ```

## Environment Variables

You can also set the `DATABASE_URL` environment variable to use a custom connection string:

```bash
# Windows
set DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable

# Linux/Mac
export DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable
```

## Troubleshooting

### Common Issues

1. **"connection refused"**: PostgreSQL service is not running
   - Start PostgreSQL service
   - Check if PostgreSQL is listening on port 5432

2. **"authentication failed"**: Wrong username/password
   - Use the default `postgres` user with password `postgres`
   - Or create a custom user as shown above

3. **"database does not exist"**: Database not created
   - Tests will automatically create tables, but the database must exist
   - Create the database manually or use the default `postgres` database

### Test Database Configuration

The test suite will automatically try these configurations in order:

1. `DATABASE_URL` environment variable (if set)
2. `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`
3. `postgres://postgres:@localhost:5432/postgres?sslmode=disable`
4. `postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable`

## Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./internal/auth/...

# Run tests with verbose output
go test -v -count=1 ./...
```

## Notes

- Tests automatically create and clean up database tables
- No Docker required - uses local PostgreSQL instance
- Tests run in silent mode to reduce noise
- Database is cleared between test runs
