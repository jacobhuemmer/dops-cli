package mcp

import (
	"context"
	"fmt"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const runbookSchema = `# Runbook YAML Schema

A runbook is defined by a ` + "`runbook.yaml`" + ` file inside a catalog directory.

## Structure

` + "```" + `
~/.dops/catalogs/<catalog-name>/<runbook-name>/
├── runbook.yaml    # Runbook definition
└── script.sh       # Automation script
` + "```" + `

## runbook.yaml

` + "```yaml" + `
name: <runbook-name>          # Must match directory name
version: 1.0.0
description: Short description of what this runbook does
risk_level: low               # low | medium | high | critical
script: script.sh             # Script filename to execute
parameters:
  - name: endpoint
    type: string              # string | boolean | integer | number | float | select | multi_select | file_path | resource_id
    required: true
    description: What this parameter does
    scope: global             # local | global | catalog | runbook
    default: ""               # Optional default value
    secret: false             # If true, value is masked and stored encrypted
    options: []               # Required for select and multi_select types
` + "```" + `

## Parameter Types

| Type | Description | Validation |
|------|-------------|------------|
| string | Free text | — |
| boolean | Yes/No toggle | — |
| integer | Whole number (negative ok) | strconv.Atoi |
| number | Non-negative whole number (0+) | Must be >= 0 |
| float | Decimal number | strconv.ParseFloat |
| select | Single choice from options | Requires options list |
| multi_select | Multiple choices from options | Requires options list |
| file_path | File system path | — |
| resource_id | Resource identifier (ARN, URI) | — |

## Risk Levels

| Level | Confirmation |
|-------|-------------|
| low | Execute immediately |
| medium | Execute immediately |
| high | Requires y/N confirmation |
| critical | Requires typing the runbook ID |

## Scopes

| Scope | Saved to |
|-------|----------|
| local | Not saved — one-time value, never prompted to save |
| global | config.json → vars.global.<name> |
| catalog | config.json → vars.catalog.<catalog>.<name> |
| runbook | config.json → vars.catalog.<catalog>.runbooks.<runbook>.<name> |
`

const shellStyleGuide = `# Shell Script Style Guide for dops Runbooks

Based on the Google Shell Style Guide with POSIX compatibility for Linux/macOS.

## File Header

` + "```sh" + `
#!/bin/sh
#
# Brief description of what the script does.
` + "```" + `

Use ` + "`#!/bin/sh`" + ` for POSIX compatibility across Linux and macOS. Avoid bash-specific features unless the runbook explicitly requires bash.

## Environment Variables

dops passes parameters as uppercase environment variables:

` + "```sh" + `
# Parameter "endpoint" becomes $ENDPOINT
# Parameter "dry_run" becomes $DRY_RUN
# Parameter "api_token" becomes $API_TOKEN
ENDPOINT="${ENDPOINT:?endpoint is required}"
DRY_RUN="${DRY_RUN:-false}"
` + "```" + `

## Error Handling

` + "```sh" + `
set -eu

# Use trap for cleanup
cleanup() {
  rm -f "${TMPFILE}"
}
trap cleanup EXIT
` + "```" + `

Note: ` + "`set -o pipefail`" + ` is not POSIX. Use ` + "`set -eu`" + ` for portable error handling.

## Functions

` + "```sh" + `
# Use lowercase with underscores. Put "main" at the bottom.
check_health() {
  url="$1"
  timeout="${2:-5}"

  if curl -sf --max-time "${timeout}" "${url}" > /dev/null; then
    echo "✓ ${url} is healthy"
  else
    echo "✗ ${url} is unhealthy" >&2
    return 1
  fi
}

main() {
  check_health "${ENDPOINT}"
}

main "$@"
` + "```" + `

## Key Rules

1. **Quote all variables**: ` + "`\"${var}\"`" + ` not ` + "`$var`" + `
2. **Use POSIX test**: ` + "`[ -n \"${var}\" ]`" + ` not ` + "`[[ ]]`" + `
3. **Use $(command)** not backticks
4. **Stderr for errors**: ` + "`echo \"error\" >&2`" + `
5. **Indent with 2 spaces**, no tabs
6. **Max line length**: 80 characters
7. **Put main() at the bottom** of the script
8. **Use set -eu** at the top (not pipefail — not POSIX)
9. **Avoid bash-isms**: no ` + "`local -r`" + `, no ` + "`[[ ]]`" + `, no ` + "`{,,}`" + ` brace expansion, no arrays
10. **Portable commands**: prefer ` + "`printf`" + ` over ` + "`echo -e`" + `, use ` + "`command -v`" + ` over ` + "`which`" + `

## Output Conventions for dops

` + "```bash" + `
# Stage headers
echo "==> Stage 1/3: Build"

# Indented details
echo "    Compiling source..."

# Success indicator
echo "✓ Build complete"

# Failure indicator
echo "✗ Build failed" >&2

# Summary block
echo "========================================="
echo "  Summary"
echo "========================================="
echo "  Status: SUCCESS"
echo "========================================="
` + "```" + `
`

// registerPrompts adds MCP prompts for runbook creation.
func (s *Server) registerPrompts() {
	s.srv.AddPrompt(
		&mcpsdk.Prompt{
			Name:        "create-runbook",
			Description: "Create a new dops runbook with the correct YAML schema and shell script template",
			Arguments: []*mcpsdk.PromptArgument{
				{
					Name:        "catalog",
					Description: "Catalog name to create the runbook in (e.g. default, infra, demo)",
					Required:    true,
				},
				{
					Name:        "name",
					Description: "Runbook name (lowercase, hyphenated, e.g. check-health)",
					Required:    true,
				},
				{
					Name:        "description",
					Description: "Short description of what the runbook does",
					Required:    true,
				},
				{
					Name:        "risk_level",
					Description: "Risk level: low, medium, high, or critical",
					Required:    false,
				},
			},
		},
		func(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
			catalog := req.Params.Arguments["catalog"]
			name := req.Params.Arguments["name"]
			description := req.Params.Arguments["description"]
			riskLevel := req.Params.Arguments["risk_level"]
			if riskLevel == "" {
				riskLevel = "low"
			}

			dopsHome := s.dopsHome
			if dopsHome == "" {
				dopsHome = "~/.dops"
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Create a new dops runbook at:\n\n"))
			sb.WriteString(fmt.Sprintf("  %s/catalogs/%s/%s/\n", dopsHome, catalog, name))
			sb.WriteString(fmt.Sprintf("  ├── runbook.yaml\n"))
			sb.WriteString(fmt.Sprintf("  └── script.sh\n\n"))
			sb.WriteString(fmt.Sprintf("## runbook.yaml\n\n"))
			sb.WriteString(fmt.Sprintf("```yaml\nname: %s\nversion: 1.0.0\ndescription: %s\nrisk_level: %s\nscript: script.sh\nparameters: []\n```\n\n", name, description, riskLevel))
			sb.WriteString("Fill in the parameters list based on what inputs the script needs.\n\n")
			sb.WriteString("## script.sh\n\n")
			sb.WriteString("Follow the Google Shell Style Guide. Use the template below as a starting point.\n\n")
			sb.WriteString("```sh\n#!/bin/sh\nset -eu\n\n# TODO: Add parameter variables\n# ENDPOINT=\"${ENDPOINT:?endpoint is required}\"\n\nmain() {\n  echo \"==> Running " + name + "\"\n  # TODO: Implement\n  echo \"✓ Done\"\n}\n\nmain \"$@\"\n```\n\n")
			sb.WriteString("Make the script executable: `chmod +x script.sh`\n")

			return &mcpsdk.GetPromptResult{
				Description: fmt.Sprintf("Create runbook %s.%s", catalog, name),
				Messages: []*mcpsdk.PromptMessage{
					{
						Role:    "user",
						Content: &mcpsdk.TextContent{Text: sb.String()},
					},
				},
			}, nil
		},
	)
}
