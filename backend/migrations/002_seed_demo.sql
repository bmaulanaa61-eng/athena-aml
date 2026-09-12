-- Demo seed data for local Docker runs.
-- Demo API key: os_test_demo_super_secret_key_please_change
-- Demo dashboard login: admin@athena.local / Admin123!

INSERT INTO tenants (id, name, status, plan)
VALUES ('00000000-0000-4000-8000-000000000001', 'PT NusantaraPay Digital Indonesia', 'ACTIVE', 'GROWTH');

INSERT INTO api_keys (id, tenant_id, name, key_prefix, key_hash, status)
VALUES (
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000001',
    'Local Demo API Key',
    'os_test_demo',
    '$2a$10$WANxs56/qlNOgMOYzaZxv.cwUY8x9ppCKt1ps4FlbXYy6Fa3xSE1K',
    'ACTIVE'
);

INSERT INTO users (id, tenant_id, email, password_hash, role, status)
VALUES (
    '00000000-0000-4000-8000-000000000201',
    '00000000-0000-4000-8000-000000000001',
    'admin@athena.local',
    '$2a$10$lNdtCyHdTl2cBZLRThFlQOZo/f8KzvB/tcJDyEVkmAuyKmO/W04y2',
    'ADMIN',
    'ACTIVE'
);

INSERT INTO aml_rules (tenant_id, rule_code, name, description, rule_type, configuration, severity, enabled)
VALUES
('00000000-0000-4000-8000-000000000001', 'SML-0001', 'Amount Spike', 'Transaction amount greater than 300% of recent average.', 'AMOUNT_THRESHOLD', '{"multiplier": 3, "weight": 35}'::jsonb, 'HIGH', TRUE),
('00000000-0000-4000-8000-000000000001', 'SML-0002', 'Velocity Structuring', 'More than 3 recent transactions in the monitoring window.', 'VELOCITY', '{"recent_count": 3, "weight": 25}'::jsonb, 'MEDIUM', TRUE),
('00000000-0000-4000-8000-000000000001', 'SML-0003', 'New Counterparty', 'First-time account-to-counterparty relation.', 'NEW_COUNTERPARTY', '{"weight": 25}'::jsonb, 'MEDIUM', TRUE),
('00000000-0000-4000-8000-000000000001', 'SML-0004', 'Off-hours Transaction', 'Transaction happens during dormant hours.', 'TEMPORAL_PATTERN', '{"start_hour": 1, "end_hour": 5, "weight": 15}'::jsonb, 'LOW', TRUE);
