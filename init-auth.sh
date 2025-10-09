#!/bin/bash
set -e

# Wait for PostgreSQL to start
sleep 2

# Set password for fleettracker user
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    ALTER USER fleettracker WITH PASSWORD 'password123';
EOSQL

# Update pg_hba.conf to allow password authentication
cat > /var/lib/postgresql/data/pg_hba.conf <<-EOF
# TYPE  DATABASE        USER            ADDRESS                 METHOD
local   all             all                                     trust
host    all             all             127.0.0.1/32            md5
host    all             all             0.0.0.0/0               md5
host    all             all             ::1/128                 md5
local   replication     all                                     trust
host    replication     all             127.0.0.1/32            md5
host    replication     all             ::1/128                 md5
EOF

# Reload PostgreSQL configuration
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "SELECT pg_reload_conf();"

echo "✅ PostgreSQL authentication configured for password md5"

