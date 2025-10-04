# Seed Data

Realistic Indonesian fleet management test data.

## Quick Start

```bash
# Seed all data
make seed

# Seed specific tables
make seed-companies
make seed-users

# Fresh start
make db-reset  # ⚠️ Deletes all data!
```

## What's Included

- **2 Companies:** PT Logistik Jakarta Raya, CV Transport Surabaya Jaya
- **5 Users:** 1 admin, 2 managers, 2 operators
- **10 Vehicles:** Indonesian license plates (B/L series)
- **5 Drivers:** Valid SIM (driver's licenses)
- **100+ GPS Tracks:** Real Jakarta & Surabaya routes
- **20 Trips:** With fuel consumption data

## Test Credentials

```
Email:    admin@logistikjkt.co.id
Password: password123
```

All users use password: `password123`

## Indonesian Data Formats

- **NPWP:** `XX.XXX.XXX.X-XXX.XXX` (Tax ID)
- **NIK:** 16-digit National ID
- **SIM:** Driver's License
- **License Plates:** B (Jakarta), L (Surabaya)
- **GPS Routes:**
  - Jakarta: Monas → Blok M (~7km)
  - Surabaya: Tugu Pahlawan → Delta Plaza (~5km)

## Features

- ✅ Idempotent (safe to run multiple times)
- ✅ 100% Indonesian-compliant formats
- ✅ Real GPS coordinates
- ✅ Realistic fuel consumption (8-10 km/L)

See main README for more details.
