package output

type OutputLineMsg struct {
	Text     string
	IsStderr bool
}

type ExecutionDoneMsg struct {
	LogPath string
	Err     error
}
