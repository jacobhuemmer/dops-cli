Output tapes/gifs/wizard.gif

Set Shell "zsh"
Set Width 1400
Set Height 900
Set FontSize 16
Set Theme "Catppuccin Mocha"

Type "DOPS_NO_ALT_SCREEN=1 ./bin/dops"
Enter
Sleep 2s

# Navigate to check-health and run
Enter
Sleep 1s

Screenshot tapes/screenshots/wizard.png

# Type endpoint param
Type "google.com"
Sleep 500ms

Screenshot tapes/screenshots/wizard-input.png

Enter
Sleep 1s

Screenshot tapes/screenshots/wizard-submit.png

Type "q"
Sleep 500ms
