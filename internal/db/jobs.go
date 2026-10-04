package db

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("job not found")

type Result struct {
	ExitCode   int
	CPUSeconds float64
	PeakMemory int64
}

func (d *DB) MarkRunning(ctx context.Context, id, pod string) error {
	tag, err := d.Pool.Exec(ctx,
		`UPDATE jobs SET status='Running', started_at=now(), pod_name=$2 WHERE id=$1`,
		id, pod)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *DB) MarkFinished(ctx context.Context, id string, r Result) error {
	status := "Complete"
	if r.ExitCode != 0 {
		status = "Failed"
	}
	_, err := d.Pool.Exec(ctx,
		`UPDATE jobs
		 SET status=$2, finished_at=now(), exit_code=$3,
		     cpu_seconds=$4, peak_memory_bytes=$5
		 WHERE id=$1`,
		id, status, r.ExitCode, r.CPUSeconds, r.PeakMemory)
	return err
}