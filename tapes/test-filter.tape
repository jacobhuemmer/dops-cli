Output tapes/screenshots/test-filter.gif

Set Shell "zsh"
Set Width 1200
Set Height 800
Set FontSize 14
Set Theme "Catppuccin Mocha"
Type "DOPS_NO_ALT_SCREEN=1 ./dops"
Enter
Sleep 2s

Screenshot tapes/screenshots/filter-1-before.png

# Activate search
Type "/"
Sleep 500ms

Screenshot tapes/screenshots/filter-2-empty.png

# Type a query
Type "drain"
Sleep 500ms

Screenshot tapes/screenshots/filter-3-query.png

# Escape to clear
Escape
Sleep 500ms

Screenshot tapes/screenshots/filter-4-cleared.png

Type "q"
Sleep 500ms
