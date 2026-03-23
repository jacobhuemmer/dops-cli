Output tapes/screenshots/test-arrows.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# Initial — cursor on default/ header
Screenshot tapes/screenshots/arrows-1-start.png

# Left arrow — collapse default/
Left
Sleep 500ms
Screenshot tapes/screenshots/arrows-2-collapse.png

# Right arrow — expand default/
Right
Sleep 500ms
Screenshot tapes/screenshots/arrows-3-expand.png

# Down to hello-world, then Left — jumps to parent
Down
Sleep 300ms
Left
Sleep 500ms
Screenshot tapes/screenshots/arrows-4-jump.png

Type "q"
Sleep 500ms
