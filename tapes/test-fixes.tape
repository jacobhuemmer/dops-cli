Output tapes/screenshots/test-fixes.png

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

Screenshot tapes/screenshots/fix-initial.png

# Navigate to hello-world
Down
Sleep 500ms

Screenshot tapes/screenshots/fix-navigate.png

# Run hello-world (has saved greeting=world)
Enter
Sleep 2s

Screenshot tapes/screenshots/fix-run.png

# Navigate to drain-node (output should clear)
Down
Down
Down
Sleep 500ms

Screenshot tapes/screenshots/fix-clear.png

Type "q"
Sleep 500ms
