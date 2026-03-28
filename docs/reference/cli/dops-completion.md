---
layout: default
title: dops completion
nav_order: 8
parent: CLI Commands
grand_parent: Reference
---

# dops completion

Generate shell completion scripts.

## Synopsis

```
dops completion <shell>
```

## Description

Generates autocompletion scripts for the specified shell. The generated script should be saved to your shell's completion directory.

Supported shells: `bash`, `zsh`, `fish`, `powershell`.

## Examples

```sh
# Bash
dops completion bash > /etc/bash_completion.d/dops

# Zsh
dops completion zsh > "${fpath[1]}/_dops"

# Fish
dops completion fish > ~/.config/fish/completions/dops.fish

# PowerShell
dops completion powershell | Out-String | Invoke-Expression
```

## See also

- [dops](dops) — launch the TUI
