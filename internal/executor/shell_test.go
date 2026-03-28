//go:build !windows

package executor

import "testing"

func TestShellFor_ShScript(t *testing.T) {
	shell, args := ShellFor("foo/bar/script.sh")
	if shell != "sh" {
		t.Errorf("shell = %q, want sh", shell)
	}
	if len(args) != 1 || args[0] != "foo/bar/script.sh" {
		t.Errorf("args = %v, want [foo/bar/script.sh]", args)
	}
}

func TestShellFor_Ps1Script(t *testing.T) {
	shell, args := ShellFor("foo/bar/deploy.ps1")
	if shell != "pwsh" {
		t.Errorf("shell = %q, want pwsh", shell)
	}
	want := []string{"-NoProfile", "-NonInteractive", "-File", "foo/bar/deploy.ps1"}
	if len(args) != len(want) {
		t.Fatalf("args length = %d, want %d", len(args), len(want))
	}
	for i, w := range want {
		if args[i] != w {
			t.Errorf("args[%d] = %q, want %q", i, args[i], w)
		}
	}
}

func TestShellFor_Ps1CaseInsensitive(t *testing.T) {
	shell, _ := ShellFor("deploy.PS1")
	if shell != "pwsh" {
		t.Errorf("shell = %q, want pwsh for .PS1", shell)
	}
}

func TestShellFor_NoExtension(t *testing.T) {
	shell, args := ShellFor("myscript")
	if shell != "sh" {
		t.Errorf("shell = %q, want sh (default)", shell)
	}
	if len(args) != 1 || args[0] != "myscript" {
		t.Errorf("args = %v, want [myscript]", args)
	}
}
