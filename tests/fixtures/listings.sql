INSERT INTO agents (id, name, email, phone, agency, created_at, updated_at) VALUES
    ('a1000000-0000-0000-0000-000000000001', 'Ada Okafor',   'ada@realty.ng',   '+2348011111111', 'Realty NG',  NOW(), NOW()),
    ('a1000000-0000-0000-0000-000000000002', 'Bola Adeyemi', 'bola@homes.ng',   '+2348022222222', 'Homes NG',   NOW(), NOW());

INSERT INTO listings (id, title, description, price, type, bedrooms, address, latitude, longitude, agent_id, created_at, updated_at) VALUES
    ('b1000000-0000-0000-0000-000000000001', '3-Bed Flat in Lekki',        'Spacious flat',    1500000, 'rent',     3, '14 Admiralty Way, Lekki Phase 1, Lagos',   6.4281,  3.4219,  'a1000000-0000-0000-0000-000000000001', NOW(), NOW()),
    ('b1000000-0000-0000-0000-000000000002', '2-Bed Apartment Victoria Island', 'Modern apartment', 2500000, 'sale',     2, '5 Adeola Odeku St, Victoria Island, Lagos', 6.4280,  3.4296,  'a1000000-0000-0000-0000-000000000001', NOW(), NOW()),
    ('b1000000-0000-0000-0000-000000000003', 'Studio Shortlet Yaba',       'Cosy studio',       80000,  'shortlet', 1, '22 Herbert Macaulay Way, Yaba, Lagos',     6.5095,  3.3711,  'a1000000-0000-0000-0000-000000000002', NOW(), NOW()),
    ('b1000000-0000-0000-0000-000000000004', '5-Bed Duplex Ikoyi',         'Luxury duplex',    8500000, 'sale',     5, '3 Bourdillon Rd, Ikoyi, Lagos',            6.4474,  3.4353,  'a1000000-0000-0000-0000-000000000002', NOW(), NOW()),
    ('b1000000-0000-0000-0000-000000000005', '4-Bed House Abuja Maitama',  'Detached house',   3200000, 'rent',     4, '12 Panama St, Maitama, Abuja',             9.0820,  7.4891,  'a1000000-0000-0000-0000-000000000001', NOW(), NOW());
