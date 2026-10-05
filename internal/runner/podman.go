package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ValidImage blocks values that podman would read as a flag.
func ValidImage(image string) error {
	if image == "" || strings.HasPrefix(image, "-") || strings.ContainsAny(image, " \t\n") {
		return fmt.Errorf("invalid image %q", image)
	}
	return nil
}

func podman(ctx context.Context, logPath string, args ...string) (int, error) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return -1, err
	}
	defer f.Close()
	out := io.MultiWriter(f, os.Stdout)

	cmd := exec.CommandContext(ctx, "podman", args...)
	cmd.Stdout, cmd.Stderr = out, out

	err = cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), nil
	}
	if err != nil {
		return -1, err
	}
	return 0, nil
}

// Pull downloads the image.
func Pull(ctx context.Context, image, logPath string) (int, error) {
	return podman(ctx, logPath, "pull", image)
}

// Run starts the image and returns its exit code.
func Run(ctx context.Context, image, logPath string) (int, error) {
	return podman(ctx, logPath, "run", "--rm", "--network=host", "--pid=host", image)
}
