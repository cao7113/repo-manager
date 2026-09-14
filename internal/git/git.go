package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Status struct {
	Branch    string
	Dirty     bool
	Untracked bool
	Conflicts bool
	Ahead     int
	Behind    int
}

func run(ctx context.Context, dir string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return strings.TrimSpace(string(output)), nil
}

func Root(ctx context.Context, path string) (string, error) {
	return run(ctx, path, "rev-parse", "--show-toplevel")
}

func RemoteURL(ctx context.Context, path string) (string, error) {
	output, err := run(ctx, path, "remote", "get-url", "origin")
	if err != nil {
		return "", nil
	}
	return output, nil
}

func Clone(ctx context.Context, args ...string) error {
	command := exec.CommandContext(ctx, "git", append([]string{"clone"}, args...)...)
	command.Stdout = nil
	command.Stderr = nil
	if output, err := command.CombinedOutput(); err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("git clone: %s", message)
	}
	return nil
}

func StatusOf(ctx context.Context, path string) (Status, error) {
	output, err := run(ctx, path, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return Status{}, err
	}
	status := Status{}
	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			fields := strings.Fields(line)
			if len(fields) == 4 {
				fmt.Sscanf(fields[2], "+%d", &status.Ahead)
				fmt.Sscanf(fields[3], "-%d", &status.Behind)
			}
		case strings.HasPrefix(line, "u "):
			status.Conflicts = true
			status.Dirty = true
		case strings.HasPrefix(line, "? "):
			status.Untracked = true
			status.Dirty = true
		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			status.Dirty = true
		}
	}
	return status, nil
}
