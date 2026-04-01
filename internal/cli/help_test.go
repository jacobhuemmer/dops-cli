package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "A test command",
		Long:  "A longer description of the test command",
	}
}

func TestHelpFunc(t *testing.T) {
	cmd := newTestCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)

	out := buf.String()
	if !strings.Contains(out, "test") {
		t.Error("output should contain command name")
	}
	if !strings.Contains(out, "A longer description") {
		t.Error("output should contain long description")
	}
}

func TestRenderHelpHeader_Root(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "dops",
		Short: "DevOps toolkit",
	}

	out := renderHelpHeader(cmd)

	if !strings.Contains(out, "do(ops) cli") {
		t.Error("root header should contain 'do(ops) cli'")
	}
	if !strings.Contains(out, "DevOps toolkit") {
		t.Error("root header should contain short description")
	}
}

func TestRenderHelpHeader_Root_LongDesc(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "dops",
		Long: "A longer root description",
	}

	out := renderHelpHeader(cmd)

	if !strings.Contains(out, "A longer root description") {
		t.Error("root header should prefer Long over Short")
	}
}

func TestRenderHelpHeader_Root_NoDesc(t *testing.T) {
	cmd := &cobra.Command{
		Use: "dops",
	}

	out := renderHelpHeader(cmd)

	if !strings.Contains(out, "do(ops) cli") {
		t.Error("root header should contain 'do(ops) cli' even without description")
	}
}

func TestRenderHelpHeader_Subcommand(t *testing.T) {
	root := &cobra.Command{Use: "dops"}
	child := &cobra.Command{
		Use:   "run",
		Short: "Run a task",
		Long:  "Run a task with full options",
	}
	root.AddCommand(child)

	out := renderHelpHeader(child)

	if !strings.Contains(out, "dops run") {
		t.Error("subcommand header should contain command path")
	}
	if !strings.Contains(out, "Run a task with full options") {
		t.Error("subcommand header should prefer Long description")
	}
}

func TestRenderHelpHeader_Subcommand_ShortOnly(t *testing.T) {
	root := &cobra.Command{Use: "dops"}
	child := &cobra.Command{
		Use:   "run",
		Short: "Run a task",
	}
	root.AddCommand(child)

	out := renderHelpHeader(child)

	if !strings.Contains(out, "Run a task") {
		t.Error("subcommand header should fall back to Short description")
	}
}

func TestRenderHelpHeader_Subcommand_NoDesc(t *testing.T) {
	root := &cobra.Command{Use: "dops"}
	child := &cobra.Command{Use: "run"}
	root.AddCommand(child)

	out := renderHelpHeader(child)

	if !strings.Contains(out, "dops run") {
		t.Error("subcommand header should contain command path")
	}
}

func TestRenderHelpUsage_Basic(t *testing.T) {
	cmd := newTestCmd()

	out := renderHelpUsage(cmd)

	if !strings.Contains(out, "USAGE") {
		t.Error("output should contain USAGE section")
	}
	if !strings.Contains(out, "test") {
		t.Error("output should contain usage line with command name")
	}
}

func TestRenderHelpUsage_WithSubcommands(t *testing.T) {
	root := &cobra.Command{Use: "dops"}
	child := &cobra.Command{
		Use:   "run",
		Short: "Run a task",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(child)

	out := renderHelpUsage(root)

	if !strings.Contains(out, "COMMANDS") {
		t.Error("output should contain COMMANDS section when subcommands exist")
	}
	if !strings.Contains(out, "run") {
		t.Error("output should list subcommand name")
	}
	if !strings.Contains(out, "Run a task") {
		t.Error("output should list subcommand description")
	}
	if !strings.Contains(out, "[command]") {
		t.Error("usage line should show [command] placeholder")
	}
	if !strings.Contains(out, "--help") {
		t.Error("output should contain help footer")
	}
}

func TestRenderHelpUsage_WithExamples(t *testing.T) {
	cmd := &cobra.Command{
		Use:     "test",
		Short:   "A test",
		Example: "test --verbose\ntest --dry-run",
	}

	out := renderHelpUsage(cmd)

	if !strings.Contains(out, "test --verbose") {
		t.Error("output should contain first example line")
	}
	if !strings.Contains(out, "test --dry-run") {
		t.Error("output should contain second example line")
	}
}

func TestRenderHelpUsage_WithFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "A test",
	}
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().String("output", "json", "Output format")

	out := renderHelpUsage(cmd)

	if !strings.Contains(out, "FLAGS") {
		t.Error("output should contain FLAGS section")
	}
	if !strings.Contains(out, "--verbose") {
		t.Error("output should contain --verbose flag")
	}
	if !strings.Contains(out, "-v") {
		t.Error("output should contain -v shorthand")
	}
	if !strings.Contains(out, "--output") {
		t.Error("output should contain --output flag")
	}
	if !strings.Contains(out, "json") {
		t.Error("output should contain default value for output flag")
	}
}

func TestRenderHelpUsage_WithInheritedFlags(t *testing.T) {
	root := &cobra.Command{Use: "dops"}
	root.PersistentFlags().Bool("debug", false, "Enable debug mode")
	child := &cobra.Command{
		Use:   "run",
		Short: "Run a task",
	}
	root.AddCommand(child)

	out := renderHelpUsage(child)

	if !strings.Contains(out, "GLOBAL FLAGS") {
		t.Error("output should contain GLOBAL FLAGS section")
	}
	if !strings.Contains(out, "--debug") {
		t.Error("output should contain inherited --debug flag")
	}
}

func TestRenderFlag_Hidden(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("secret", false, "Hidden flag")
	_ = cmd.Flags().MarkHidden("secret")

	out := renderHelpUsage(cmd)

	if strings.Contains(out, "secret") {
		t.Error("output should not contain hidden flags")
	}
}

func TestRenderCommands_SkipsHiddenAndDeprecated(t *testing.T) {
	root := &cobra.Command{Use: "dops"}
	noop := func(cmd *cobra.Command, args []string) error { return nil }
	visible := &cobra.Command{Use: "run", Short: "Run a task", RunE: noop}
	hidden := &cobra.Command{Use: "internal", Short: "Hidden", Hidden: true, RunE: noop}
	deprecated := &cobra.Command{Use: "old", Short: "Old", Deprecated: "use run", RunE: noop}
	root.AddCommand(visible, hidden, deprecated)

	var buf bytes.Buffer
	renderCommands(&buf, root)
	out := buf.String()

	if !strings.Contains(out, "run") {
		t.Error("output should contain visible command")
	}
	if strings.Contains(out, "internal") {
		t.Error("output should not contain hidden command")
	}
	if strings.Contains(out, "old") {
		t.Error("output should not contain deprecated command")
	}
}
