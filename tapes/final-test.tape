Output tapes/screenshots/final-test.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# Initial view — check-health selected
Screenshot tapes/screenshots/final-initial.png

# Navigate to list-pods
Down
Down
Sleep 500ms

# Run list-pods (has saved params)
Enter
Sleep 2s

Screenshot tapes/screenshots/final-listpods.png

# Navigate to drain-node
Down
Sleep 500ms

Screenshot tapes/screenshots/final-drain-selected.png

# Run drain-node (has saved node_name)
Enter
Sleep 3s

Screenshot tapes/screenshots/final-drain-output.png

Type "q"
Sleep 500ms
