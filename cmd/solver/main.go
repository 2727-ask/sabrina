package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/2727-ask/jobstar/internal/config"
	"github.com/2727-ask/jobstar/internal/db"
	kube "github.com/2727-ask/jobstar/internal/kube"
	"github.com/2727-ask/jobstar/internal/model"
	"github.com/2727-ask/jobstar/internal/queue"
	"github.com/2727-ask/jobstar/internal/runner"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer closeWithTimeout("postgres", database.Close, 5*time.Second)

	taker, err := queue.NewConsumer(cfg.AMQPURL)
	if err != nil {
		return err
	}
	var closeOnce sync.Once
	closeTaker := func() {
		closeOnce.Do(func() {
			closeWithTimeout("rabbitmq", taker.Close, 5*time.Second)
		})
	}
	defer closeTaker()

	msg, ok, err := taker.Consume("jobs")
	if err != nil {
		return err
	}
	if !ok {
		log.Println("queue is empty, exiting")
		return nil
	}

	var job model.Job
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		_ = msg.Reject(false) // bad message, drop it
		return err
	}

	pod := os.Getenv("POD_NAME")
	if err := markRunning(ctx, database, job.ID, pod); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			_ = msg.Reject(false) // no row for this message, drop it
			return err
		}
		_ = msg.Nack(false, true) // DB problem, put it back
		return err
	}
	_ = msg.Ack(false)

	// The message is acked, so free the broker connection before a long job.
	closeTaker()

	if err := kube.LabelPod(ctx, job.ID); err != nil {
		log.Printf("label pod: %v", err) // not fatal
	}

	res := execute(ctx, job)

	// Use a fresh context, because ctx may be cancelled by SIGTERM.
	fctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return database.MarkFinished(fctx, job.ID, res)
}

// closeWithTimeout stops a slow Close from blocking exit.
func closeWithTimeout(name string, closeFn func(), d time.Duration) {
	done := make(chan struct{})
	go func() {
		closeFn()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(d):
		log.Printf("%s close timed out", name)
	}
}

// markRunning retries briefly. The API may publish before its commit lands.
func markRunning(ctx context.Context, d *db.DB, id, pod string) error {
	var err error
	for i := 0; i < 5; i++ {
		if err = d.MarkRunning(ctx, id, pod); !errors.Is(err, db.ErrNotFound) {
			return err
		}
		time.Sleep(time.Second)
	}
	return err
}

func execute(ctx context.Context, job model.Job) db.Result {
	if err := runner.ValidImage(job.Image); err != nil {
		log.Println(err)
		return db.Result{ExitCode: -1}
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/var/log/job"
	}

	cpuBefore := runner.CPUSeconds()
	code, err := runner.Run(ctx, job.Image, filepath.Join(logDir, "job.log"))
	if err != nil {
		log.Printf("run: %v", err)
		code = -1
	}

	return db.Result{
		ExitCode:   code,
		CPUSeconds: runner.CPUSeconds() - cpuBefore,
		PeakMemory: runner.PeakMemory(),
	}
}
