---
name: charm-styling
description: Lip Gloss v2 styling and layout patterns for Go. Use when creating or editing terminal styles, colors, borders, layouts, or visual components using Lip Gloss. Triggers on lipgloss.NewStyle, lipgloss.Color, JoinHorizontal, JoinVertical, Place, border usage, or any visual styling of terminal output.
user-invocable: false
---

# Lip Gloss v2 Styling Patterns

Import: `charm.land/lipgloss/v2`

## Creating Styles

Styles are value types (assignment copies). `Copy()` is deprecated.

```go
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(1, 2).
    Width(40)

output := style.Render("Hello")
```

Every setter returns a new `Style`. Chain freely.

## Colors (v2 Change)

Colors must use `lipgloss.Color()` — bare strings no longer work:

```go
lipgloss.Color("#FF00FF")     // hex -> TrueColor
lipgloss.Color("205")         // ANSI 256
lipgloss.Color("1")           // ANSI 16
lipgloss.NoColor{}            // transparent
```

Named constants: `lipgloss.Red`, `lipgloss.Green`, `lipgloss.BrightCyan`, etc.

## Dark/Light Background (v2 Change)

`AdaptiveColor` is removed. Use `LightDark`:

```go
hasDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
lightDark := lipgloss.LightDark(hasDark)
fg := lightDark(lipgloss.Color("#333"), lipgloss.Color("#ccc"))
```

In BubbleTea: handle `tea.BackgroundColorMsg` on startup for detection.

## Printing (v2 Change — Critical)

v2 does NOT auto-downsample colors. Outside BubbleTea, use lipgloss print functions:

```go
lipgloss.Println(style.Render("text"))   // auto-downsamples
lipgloss.Printf("Status: %s\n", styled)
lipgloss.Sprint(styled)                   // returns downsampled string
```

Inside BubbleTea, rendering handles downsampling automatically.

## Layout

```go
// Side by side
lipgloss.JoinHorizontal(lipgloss.Top, left, right)
lipgloss.JoinHorizontal(lipgloss.Center, col1, col2, col3)

// Stacked
lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
lipgloss.JoinVertical(lipgloss.Center, title, content)

// Place in a box
lipgloss.Place(80, 24, lipgloss.Center, lipgloss.Center, content)
lipgloss.PlaceHorizontal(80, lipgloss.Right, text)
lipgloss.PlaceVertical(24, lipgloss.Bottom, text)
```

Position values: `lipgloss.Top`, `lipgloss.Bottom`, `lipgloss.Left`, `lipgloss.Right`, `lipgloss.Center`.

## Borders

```go
style := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#874BFD")).
    Padding(1, 2)
```

Built-in borders: `NormalBorder()`, `RoundedBorder()`, `ThickBorder()`, `DoubleBorder()`, `BlockBorder()`, `HiddenBorder()`, `ASCIIBorder()`.

Per-side control:
```go
.BorderTop(true).BorderBottom(true).BorderLeft(false).BorderRight(false)
.BorderTopForeground(lipgloss.Color("#F00"))
```

## Dimensions

```go
.Width(40)        // min width (pads with spaces)
.Height(10)       // min height (pads with newlines)
.MaxWidth(80)     // truncates/wraps
.MaxHeight(20)    // truncates
```

## Padding and Margin (CSS-like)

```go
.Padding(1)          // all sides
.Padding(1, 2)       // top/bottom, left/right
.Padding(1, 2, 3, 4) // top, right, bottom, left (clockwise)
.PaddingLeft(4)       // individual side

.Margin(1, 2)
.MarginBackground(lipgloss.Color("#333"))
```

## Frame Size Helpers

```go
hFrame := style.GetHorizontalFrameSize()  // border + padding + margin
vFrame := style.GetVerticalFrameSize()

// Responsive content width:
contentWidth := windowWidth - hFrame
```

## Text Measurement

```go
w := lipgloss.Width(rendered)     // ANSI-aware cell width
h := lipgloss.Height(rendered)
w, h := lipgloss.Size(rendered)
```

## New v2 Features

Hyperlinks:
```go
.Hyperlink("https://example.com")
```

Underline styles:
```go
.UnderlineStyle(lipgloss.UnderlineCurly)
.UnderlineColor(lipgloss.Color("#FF0000"))
```

Color manipulation:
```go
lipgloss.Darken(c, 0.2)
lipgloss.Lighten(c, 0.3)
lipgloss.Complementary(c)
```

Gradients:
```go
lipgloss.Blend1D(steps, color1, color2, color3)
```

Style ranges:
```go
lipgloss.StyleRanges(text,
    lipgloss.NewRange(0, 5, boldStyle),
    lipgloss.NewRange(6, 11, dimStyle),
)
```

## Table, List, Tree

```go
import "charm.land/lipgloss/v2/table"
import "charm.land/lipgloss/v2/list"
import "charm.land/lipgloss/v2/tree"
```

Table:
```go
t := table.New().
    Headers("Name", "Status").
    Row("prod-01", "running").
    Row("prod-02", "stopped").
    Border(lipgloss.RoundedBorder()).
    Width(60).
    StyleFunc(func(row, col int) lipgloss.Style {
        if row == table.HeaderRow {
            return lipgloss.NewStyle().Bold(true)
        }
        return lipgloss.NewStyle()
    })
fmt.Println(t.Render())
```

List:
```go
l := list.New("Item 1", "Item 2", "Item 3").
    Enumerator(list.Bullet)
```

Tree:
```go
t := tree.Root("Root").
    Child("Foo", tree.Root("Bar").Child("Baz"))
```

## Style Organization Pattern

Keep styles in a dedicated file (e.g., `styles.go`):

```go
package tui

var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FAFAFA")).
        MarginBottom(1)

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000"))

    subtleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#666666"))
)
```

Do NOT scatter style definitions across View functions. Centralize them for consistency.

## Common Mistakes to Avoid

- Do NOT use bare color strings — `lipgloss.Color("#fff")` not `"#fff"`
- Do NOT use `fmt.Println` with styled output outside BubbleTea — use `lipgloss.Println`
- Do NOT use `style.Copy()` — just assign: `newStyle := style.Bold(true)`
- Do NOT use `AdaptiveColor` — it's removed in v2, use `LightDark`
- Do NOT hardcode widths — use `WindowSizeMsg` and compute responsive sizes
