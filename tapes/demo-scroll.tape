Output tapes/gifs/scroll.gif

Set Shell "zsh"
Set Width 1400
Set Height 900
Set FontSize 16
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./bin/dops"
Enter
Sleep 2s

# Run check-health
Enter
Sleep 1s
Type "google.com"
Sleep 300ms
Enter
Sleep 4s

# Tab to output, scroll
Tab
Sleep 300ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms

Screenshot tapes/screenshots/scroll.png

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

Screenshot tapes/screenshots/scroll-bottom.png

Type "q"
Sleep 500ms
