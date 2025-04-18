ALTER TABLE pvz 
ALTER COLUMN id SET DEFAULT gen_random_uuid(),
ALTER COLUMN registration_date SET DEFAULT NOW();