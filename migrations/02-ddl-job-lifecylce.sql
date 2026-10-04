ALTER TABLE jobs
  ADD COLUMN started_at  TIMESTAMPTZ,
  ADD COLUMN finished_at TIMESTAMPTZ,
  ADD COLUMN exit_code   INT,
  ADD COLUMN cpu_seconds DOUBLE PRECISION,
  ADD COLUMN peak_memory_bytes BIGINT;

ALTER TABLE jobs DROP CONSTRAINT jobs_status_check;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check
  CHECK (status IN ('Pending','Running','Complete','Failed'));