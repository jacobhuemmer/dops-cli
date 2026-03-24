Output tapes/screenshots/full-flow.gif

Set Shell "zsh"
Set Width 1400
Set Height 900
Set FontSize 16
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s
Screenshot tapes/screenshots/flow-01-startup.png

# Execute check-health
Enter
Sleep 1s
Screenshot tapes/screenshots/flow-02-wizard.png

Type "google.com"
Sleep 300ms
Enter
Sleep 3s
Screenshot tapes/screenshots/flow-03-output.png

# Tab to output pane
Tab
Sleep 300ms
Screenshot tapes/screenshots/flow-04-focused.png

# Scroll up
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 300ms
Screenshot tapes/screenshots/flow-05-scrolled-up.png

# Scroll back down
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Screenshot tapes/screenshots/flow-06-scrolled-down.png

# Tab back to sidebar, navigate to a different runbook
Tab
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Screenshot tapes/screenshots/flow-07-sidebar-nav.png

Type "q"
Sleep 500ms
