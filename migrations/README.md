# Database Migrations

SQL-based migrations using [golang-migrate](https://github.com/golang-migrate/migrate).

## Quick Start

```bash
# Apply migrations
make migrate-up

# Rollback
make migrate-down

# Check version
make migrate-version

# Create new migration
make migrate-create NAME=add_feature
```

## File Format

```
{version}_{description}.{up|down}.sql
```

Example:
- `001_initial_schema.up.sql` - Creates tables
- `001_initial_schema.down.sql` - Drops tables

## Best Practices

1. **Idempotent:** Use `IF EXISTS` / `IF NOT EXISTS`
2. **Reversible:** Every `.up.sql` needs a `.down.sql`
3. **Testable:** Always test rollback
4. **Small:** One logical change per migration

## Troubleshooting

**Dirty database:**
```bash
make migrate-force VERSION=1
make migrate-up
```

**Connection refused:**
```bash
make docker-up
# Wait 30 seconds
make migrate-up
```

See main README for more details.
