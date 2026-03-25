Output tapes/screenshots/themes.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"

# Set catppuccin-mocha theme
Type "DOPS_NO_ALT_SCREEN=1 ./dops config set theme=catppuccin-mocha"
Enter
Sleep 500ms

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s
Screenshot tapes/screenshots/theme-catppuccin-mocha.png

Type "q"
Sleep 500ms

# Set nord theme
Type "DOPS_NO_ALT_SCREEN=1 ./dops config set theme=nord"
Enter
Sleep 500ms

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s
Screenshot tapes/screenshots/theme-nord.png

Type "q"
Sleep 500ms

# Set rosepine-dawn
Type "DOPS_NO_ALT_SCREEN=1 ./dops config set theme=rosepine-dawn"
Enter
Sleep 500ms

Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s
Screenshot tapes/screenshots/theme-rosepine-dawn.png

Type "q"
Sleep 500ms

# Reset to tokyonight
Type "DOPS_NO_ALT_SCREEN=1 ./dops config set theme=tokyonight"
Enter
Sleep 500ms
