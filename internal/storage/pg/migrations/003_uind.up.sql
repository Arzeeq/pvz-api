CREATE UNIQUE INDEX IF NOT EXISTS receptions_one_in_progress_per_pvz 
ON receptions (pvz_id) 
WHERE status = 'in_progress';