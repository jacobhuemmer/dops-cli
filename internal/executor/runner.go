package executor

import "context"

type OutputLine struct {
	Text     string
	IsStderr bool
}

type Runner interface {
	Run(ctx context.Context, scriptPath string, env map[string]string) (<-chan OutputLine, <-chan error)
}
