Output tapes/screenshots/test-current.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 3s

Screenshot tapes/screenshots/current-initial.png

# Navigate down a few times
Down
Down
Down
Down
Sleep 500ms

Screenshot tapes/screenshots/current-nav.png

Type "q"
Sleep 500ms
