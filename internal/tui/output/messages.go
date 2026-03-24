package output

type OutputLineMsg struct {
	Text     string
	IsStderr bool
}

type ExecutionDoneMsg struct {
	LogPath string
	Err     error
}

type CopiedHeaderFlashMsg struct{}
type CopiedFooterFlashMsg struct{}
type CopyFlashExpiredMsg struct{}
