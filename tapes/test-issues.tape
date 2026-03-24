Output tapes/screenshots/issues.gif

Set Shell "zsh"
Set Width 1400
Set Height 900
Set FontSize 16
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s
Screenshot tapes/screenshots/issues-01-startup.png

# Run check-health
Enter
Sleep 1s
Type "google.com"
Sleep 300ms
Enter
Sleep 4s
Screenshot tapes/screenshots/issues-02-output.png

# Tab to output, scroll up
Tab
Sleep 300ms
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
Screenshot tapes/screenshots/issues-03-scrolled.png

# Scroll all the way down
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Screenshot tapes/screenshots/issues-04-bottom.png

Type "q"
Sleep 500ms
