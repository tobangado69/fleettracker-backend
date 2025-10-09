-- Quick Seed Data for FleetTracker Pro
-- Run this in pgAdmin Query Tool

-- Insert Companies
INSERT INTO companies (name, email, phone, npwp, city, province, country, company_type, fleet_size, max_vehicles, is_active)
VALUES 
  ('PT Fleet Indonesia', 'contact@fleet.id', '+62 21 5551234', '01.234.567.8-901.000', 'Jakarta', 'DKI Jakarta', 'Indonesia', 'PT', 10, 100, true),
  ('CV TransJaya Surabaya', 'info@transjaya.id', '+62 31 5556789', '02.345.678.9-012.000', 'Surabaya', 'Jawa Timur', 'Indonesia', 'CV', 5, 50, true)
ON CONFLICT (email) DO NOTHING;

-- Get company IDs
DO $$
DECLARE
  company1_id UUID;
  company2_id UUID;
BEGIN
  SELECT id INTO company1_id FROM companies WHERE email = 'contact@fleet.id';
  SELECT id INTO company2_id FROM companies WHERE email = 'info@transjaya.id';

  -- Insert Users
  INSERT INTO users (company_id, email, username, password, first_name, last_name, role, phone, is_active, is_verified)
  VALUES
    (company1_id, 'admin@fleet.id', 'admin_fleet', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5lW5h8TQz5yPW', 'Admin', 'Fleet', 'admin', '+62 811 1111 1111', true, true),
    (company1_id, 'manager@fleet.id', 'manager_fleet', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5lW5h8TQz5yPW', 'Manager', 'Jakarta', 'manager', '+62 812 2222 2222', true, true),
    (company2_id, 'admin@transjaya.id', 'admin_trans', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5lW5h8TQz5yPW', 'Admin', 'TransJaya', 'admin', '+62 813 3333 3333', true, true)
  ON CONFLICT (email) DO NOTHING;

  RAISE NOTICE '✅ Seed data inserted successfully!';
  RAISE NOTICE 'Default password for all users: password123';
END $$;

