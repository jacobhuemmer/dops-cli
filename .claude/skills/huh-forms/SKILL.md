---
name: huh-forms
description: Huh v2 form building patterns for Go. Use when creating interactive forms, prompts, field validation, select menus, text inputs, confirmations, or file pickers using the Huh library. Triggers on huh.NewForm, huh.NewInput, huh.NewSelect, huh.NewConfirm, and related Huh types.
user-invocable: false
---

# Huh v2 Form Patterns

Import: `charm.land/huh/v2`

## Form Structure

Forms contain Groups (pages). Groups contain Fields. One group is shown at a time.

```go
var name string
var env string
var confirm bool

form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("Deployment name").
            Value(&name).
            Validate(huh.ValidateNotEmpty()),
        huh.NewSelect[string]().
            Title("Environment").
            Options(
                huh.NewOption("Production", "prod"),
                huh.NewOption("Staging", "staging"),
                huh.NewOption("Development", "dev"),
            ).
            Value(&env),
    ),
    huh.NewGroup(
        huh.NewConfirm().
            Title("Deploy?").
            Value(&confirm),
    ),
)

err := form.Run()
```

## Field Types

**Input** — single-line text:
```go
huh.NewInput().
    Title("Name").
    Placeholder("enter name...").
    CharLimit(64).
    EchoMode(huh.EchoModePassword).  // or EchoModeNormal, EchoModeNone
    Suggestions([]string{"prod", "staging"}).
    Value(&s).
    Validate(huh.ValidateNotEmpty())
```

**Text** — multi-line text:
```go
huh.NewText().
    Title("Description").
    Lines(5).
    ShowLineNumbers(true).
    CharLimit(500).
    Value(&s)
```

**Select[T]** — single selection:
```go
huh.NewSelect[string]().
    Title("Region").
    Options(
        huh.NewOption("US East", "us-east-1"),
        huh.NewOption("EU West", "eu-west-1"),
    ).
    Height(5).      // visible option count
    Filtering(true). // enable type-to-filter
    Value(&region)
```

**MultiSelect[T]** — multiple selection:
```go
huh.NewMultiSelect[string]().
    Title("Features").
    Options(
        huh.NewOption("Logging", "logging").Selected(true),  // pre-selected
        huh.NewOption("Metrics", "metrics"),
        huh.NewOption("Tracing", "tracing"),
    ).
    Limit(2).  // max selections
    Value(&features)
```

**Confirm** — yes/no:
```go
huh.NewConfirm().
    Title("Continue?").
    Affirmative("Yes!").
    Negative("No").
    Value(&ok)
```

**Note** — display-only:
```go
huh.NewNote().
    Title("Warning").
    Description("This will affect production.").
    Next(true).
    NextLabel("I understand")
```

**FilePicker** — file selection:
```go
huh.NewFilePicker().
    Title("Select config").
    CurrentDirectory(".").
    AllowedTypes([]string{".yaml", ".yml"}).
    Value(&path)
```

## Built-in Validators

```go
huh.ValidateNotEmpty()
huh.ValidateMinLength(3)
huh.ValidateMaxLength(64)
huh.ValidateLength(3, 64)
huh.ValidateOneOf("prod", "staging", "dev")
```

Custom validation:
```go
.Validate(func(s string) error {
    if !strings.HasPrefix(s, "deploy-") {
        return fmt.Errorf("must start with 'deploy-'")
    }
    return nil
})
```

## Dynamic Forms

Use `*Func` variants when field content depends on other answers:

```go
var category string

huh.NewSelect[string]().
    Title("Subcategory").
    OptionsFunc(func() []huh.Option[string] {
        return getSubcategories(category)  // re-evaluated when &category changes
    }, &category)
```

Also available: `TitleFunc`, `DescriptionFunc`, `PlaceholderFunc`, `SuggestionsFunc`.

## Conditional Groups

Skip groups based on prior answers:

```go
huh.NewGroup(
    huh.NewInput().Title("API Key").Value(&apiKey),
).WithHideFunc(func() bool {
    return env != "prod"  // only show for production
})
```

## Form Options

```go
form.WithTheme(huh.ThemeCharm(isDark))   // v2: pass isDark bool
form.WithAccessible(true)                 // screen reader mode
form.WithWidth(60).WithHeight(20)
form.WithShowHelp(true)
form.WithShowErrors(true)
form.WithTimeout(30 * time.Second)
form.WithLayout(huh.LayoutColumns(2))     // multi-column layout
```

Built-in themes: `ThemeCharm`, `ThemeDracula`, `ThemeCatppuccin`, `ThemeBase16`, `ThemeBase`.

## Data Retrieval by Key

```go
huh.NewInput().Title("Name").Key("name").Value(&name)

// After form completes:
form.GetString("name")
form.GetInt("count")
form.GetBool("confirm")
form.Get("key")  // any
```

## BubbleTea Integration

`*huh.Form` implements `tea.Model` — embed it in a larger BubbleTea app:

```go
type model struct {
    form *huh.Form
}

func (m model) Init() tea.Cmd {
    return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    form, cmd := m.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        m.form = f
    }
    if m.form.State == huh.StateCompleted {
        // transition to next view
    }
    return m, cmd
}
```

Form states: `huh.StateNormal`, `huh.StateCompleted`, `huh.StateAborted`.
Errors: `huh.ErrUserAborted`, `huh.ErrTimeout`.

## Spinner

```go
import "charm.land/huh/v2/spinner"

err := spinner.New().
    Title("Deploying...").
    Type(spinner.Dots).
    Action(func() {
        // synchronous work
    }).
    Run()
```

## Common Mistakes to Avoid

- Do NOT pass theme by pointer in v2 — themes are passed by value
- Do NOT forget to check `form.State` for abort — user may press Escape/ctrl+c
- Do NOT put validation logic in the command handler — put it on the field with `.Validate()`
- Do NOT use `form.Run()` inside a BubbleTea program — embed the form as a `tea.Model` instead
