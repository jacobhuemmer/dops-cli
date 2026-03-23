Output tapes/screenshots/test-layout-fix.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 3s

Screenshot tapes/screenshots/layout-fix-initial.png

Down
Down
Down
Down
Sleep 500ms

Screenshot tapes/screenshots/layout-fix-nav.png

Enter
Sleep 2s

Screenshot tapes/screenshots/layout-fix-exec.png

Type "q"
Sleep 500ms
