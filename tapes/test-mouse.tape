Output tapes/screenshots/test-mouse.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

# Initial state
Screenshot tapes/screenshots/mouse-1-initial.png

# Click on drain-node (should be around row 7, col 10 in the sidebar)
Click 80 168
Sleep 500ms

Screenshot tapes/screenshots/mouse-2-click-runbook.png

# Click on default/ header to collapse (row 2, col 5)
Click 60 62
Sleep 500ms

Screenshot tapes/screenshots/mouse-3-collapse.png

# Click default/ again to expand
Click 60 62
Sleep 500ms

Screenshot tapes/screenshots/mouse-4-expand.png

Type "q"
Sleep 500ms
