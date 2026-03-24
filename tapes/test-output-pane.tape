Output tapes/screenshots/output-pane.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"

# Launch TUI
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# Screenshot: initial state — output container should be empty shell, no bg fill
Screenshot tapes/screenshots/output-initial.png

# Navigate to a runbook and execute
Down
Sleep 300ms
Down
Sleep 300ms
Screenshot tapes/screenshots/output-selected.png

# Execute the runbook
Enter
Sleep 1s

# If wizard appears, submit it
Enter
Sleep 3s

# Screenshot: after execution — should show header/log/footer sections
Screenshot tapes/screenshots/output-executed.png

# Quit
Type "q"
Sleep 500ms
