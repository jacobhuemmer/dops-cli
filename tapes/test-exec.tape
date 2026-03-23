Output tapes/screenshots/test-exec.png

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# First runbook is check-health, it needs endpoint param — should open wizard
# Let's navigate to list-pods which has saved params
Down
Down
Sleep 500ms

Screenshot tapes/screenshots/exec-before.png

Enter
Sleep 3s

Screenshot tapes/screenshots/exec-after.png

Escape
Sleep 500ms
Type "q"
Sleep 500ms
