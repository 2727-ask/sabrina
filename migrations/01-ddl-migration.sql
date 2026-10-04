CREATE TABLE IF NOT EXISTS jobs (
    id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name   TEXT NOT NULL,
    image  TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Pending'
           CHECK (status IN ('Pending', 'Running', 'Complete')),
    date   TIMESTAMPTZ NOT NULL DEFAULT now()
);