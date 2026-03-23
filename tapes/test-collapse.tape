Output tapes/screenshots/test-collapse.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# Initial — all expanded, cursor on default/ header
Screenshot tapes/screenshots/collapse-1-initial.png

# Press Enter to collapse default/
Enter
Sleep 500ms

Screenshot tapes/screenshots/collapse-2-collapsed.png

# Navigate down to local/ and collapse it too
Down
Sleep 300ms
Enter
Sleep 500ms

Screenshot tapes/screenshots/collapse-3-both.png

# Expand default/ back
Up
Sleep 300ms
Enter
Sleep 500ms

Screenshot tapes/screenshots/collapse-4-expand.png

# Navigate to hello-world and press Enter to execute
Down
Sleep 300ms
Enter
Sleep 2s

Screenshot tapes/screenshots/collapse-5-execute.png

Type "q"
Sleep 500ms
