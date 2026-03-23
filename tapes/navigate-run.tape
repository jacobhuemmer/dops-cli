Output tapes/screenshots/navigate-run.png

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "cd /Users/mason/git/dops && ./dops"
Enter
Sleep 2s

# Navigate down to hello-world
Down
Sleep 500ms

Screenshot tapes/screenshots/navigate-select.png

# Press enter to run (hello-world has required param 'greeting' missing — should open wizard)
Enter
Sleep 2s

Screenshot tapes/screenshots/navigate-run.png

Type "q"
Sleep 500ms
