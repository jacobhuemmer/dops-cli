Output tapes/screenshots/check-health.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# check-health is the first runbook, already selected. Execute it.
Enter
Sleep 1s

# Wizard: type endpoint param and submit
Type "google.com"
Sleep 300ms
Enter
Sleep 3s

Screenshot tapes/screenshots/check-health-output.png

# Tab to focus the output pane, then scroll
Tab
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Screenshot tapes/screenshots/check-health-scrolled.png

Type "q"
Sleep 500ms
