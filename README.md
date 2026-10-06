# missionctl-core

Shared Bubble Tea / Lipgloss building blocks for the [missionctl](https://github.com/aeon022/missionctl) suite.

## Why this exists

mailctl, calctl, taskctl, notectl, budgetctl, timectl and diaryctl had already
converged on an identical color palette and near-identical help-overlay code
by copy-paste — including a bug (a divider color too dark to see on many dark
terminal themes) that had to be fixed independently in all seven repos before
this package was actually adopted. All seven now depend on `theme` for their
shared palette, replacing the duplicated literals. habctl keeps its own
distinct palette on purpose and is not expected to depend on this package;
postctl has its own separate design language and is likewise out of scope.

## Packages

- `theme` — the shared `AdaptiveColor` palette (Blue/Green/Red/Amber/Muted/
  Subtle) and base Lipgloss styles used across the suite's TUIs, plus
  `NewSpinner` for the MiniDot sync/AI-call spinner every tool already builds
  the same way. Tools on Bubble Tea v2 use the resolved mirrors instead:
  `BlueV2`, `GreenV2`, `RedV2`, `AmberV2`, `MutedV2`, `SubtleV2`,
  `SelectedBgV2`, `SelectedFgV2`, `HoverBgV2`, `OnAccentV2` (`color.Color`)
  and `HoverV2` (a `lipgloss/v2` style). v2 has no `AdaptiveColor`; these are
  resolved once at startup for light vs. dark terminals.
- `keymap` — a small builder for the `?` help overlay (replacing each tool's
  hand-rolled `key()/row()/section()` helpers), plus the standard single-key
  bindings (`SearchKey "/"`, `HelpKey "?"`, `DeleteKey "d"`, `ConfirmKey "y"`,
  `QuitKey "q"`, `BackKey "esc"`, `SyncKey "s"`) already followed by every
  tool except habctl's `s` (AI-suggest, not sync — deliberate, habctl has no
  external data source to sync from).
- `config` — `DataDir(tool)` for the `~/.local/share/<tool>` convention
  (mailctl and taskctl keep using `~/Library/Application Support/<tool>` for
  their SQLite DB on purpose; this package doesn't override that).
  `ResolveDir(tool, override)` returns `override` (expanded and created) if
  the user configured one, else falls back to `DataDir(tool)` — the entry
  point for pointing a tool's data at a folder of the user's own choosing,
  see `syncdir` below. `Store` is the suite's small YAML+env settings store
  (replaced viper) — see [config.Store](#configstore) below.
- `syncdir` — makes it safe to point a tool's SQLite database at a
  user-owned synced folder (iCloud Drive, Dropbox, Syncthing, ...) instead
  of its private default directory: `JournalMode(shared)` switches WAL to
  rollback-journal the moment a directory is user-configured, since WAL's
  multi-file on-disk state (`.db`/`.db-wal`/`.db-shm`) has no cross-file
  sync-atomicity guarantee from a folder-sync client; `Acquire`/`Release`
  is a same-machine advisory flock so two processes never write at once;
  `ICloudPlaceholder` detects an "Optimize Mac Storage" stub so a tool can
  report a clear message instead of a bare missing file. Same-machine only
  — it does not and cannot solve true simultaneous cross-machine writes;
  no folder-sync client offers a cross-machine lock to build on.
- `overlay` — `Center(background, popup, width, height, inset)` composites a
  popup on top of already-rendered content instead of a full view-state
  switch replacing the whole screen (e.g. a transient `?` help panel that
  keeps the list visible around it). Uses `ansi.Cut` for ANSI-safe column
  slicing. Pass `inset > 0` when background is itself a fully bordered
  panel, so the popup can't collide with the background's own border ring —
  first use of this (habctl's help overlay) got that wrong initially and
  produced visibly doubled-up "╭──╭──╮──╮" corners before the inset clamp
  was added. `CenterDim` is the same call with the background dimmed
  (see [overlay.CenterDim](#overlaycenterdim)).

## Adoption

Migration is **incremental, not a big-bang rewrite**: a tool adopts
`missionctl-core` the next time its TUI is touched for another reason, by
adding the dependency and replacing its local color vars / help-overlay
helpers with calls into this package. No tool is required to migrate on any
particular timeline.

Status (checked against the tools' imports): `theme`, `keymap`, `overlay`,
`palette`, `config` and `syncdir` are used by almost every tool; `statusbar`
and `emptystate` by the eight TUIs that adopted them (not postctl, which has
its own layout system); `tuitest` by every tool's TUI tests; `applescript` by
calctl, mailctl, notectl, taskctl and missionctl.

```go
import (
    "github.com/aeon022/missionctl-core/theme"
    "github.com/aeon022/missionctl-core/keymap"
)

help := keymap.New("taskctl", "tasks from the terminal").
    Section("Navigation").
    Row("j / k", "move down / up").
    Row("/", "search (esc clears)").
    Section("Other").
    Row("?", "toggle this help").
    Row("q", "quit").
    String()
```

## Newer packages

All snippets below use the real exported API of this repository.

### config.Store

A tiny YAML-backed key/value store, one per tool. Keys are case-insensitive
and may be dotted. A value resolves from the environment first
(`<PREFIX>_<KEY>`, only for top-level keys), then from what was set or read.
It replaced the global `viper` instance in budgetctl, calctl, mailctl and
taskctl. Not safe for concurrent use.

```go
import coreconfig "github.com/aeon022/missionctl-core/config"

var settings = coreconfig.NewStore("config") // looks for config.yaml / config.yml

func Load() error {
    settings.SetEnvPrefix("MAILCTL") // env overrides are OFF until this is called
    settings.AddPath(filepath.Join(home, ".config", "mailctl")) // $VARS are expanded
    settings.SetDefault("sync_days", 30) // only sets the key if nothing has yet
    if err := settings.Read(); err != nil && !errors.Is(err, coreconfig.ErrNotFound) {
        return err // no file at all is fine: defaults apply
    }
    var cfg struct {
        SyncDays int `yaml:"sync_days"`
    }
    return settings.Unmarshal(&cfg) // env overrides applied; decodes via yaml tags
}

settings.Set("profiles.work.data_dir", "~/Dropbox/work") // dotted keys create nested maps
settings.GetString("profiles.work.data_dir")
settings.GetMap("profiles") // a copy: safe to delete from, then Set back
settings.Write(path)        // persists the state, WITHOUT env overrides (0600)
settings.Reset()            // clears everything incl. the env prefix — use in tests
```

The env prefix is opt-in on purpose: a test that never calls `Load()` must not
be redirected by `BUDGETCTL_DATA_DIR` from the developer's shell.

### applescript

The one place the suite shells out to `osascript`, so error handling and
escaping are identical everywhere.

```go
import "github.com/aeon022/missionctl-core/applescript"

script := fmt.Sprintf(`tell application "Notes" to make new note with properties {name:"%s"}`,
    applescript.Escape(title)) // escapes \ and "
out, err := applescript.Run(script)     // stdout, whitespace-trimmed
body, err := applescript.RunRaw(script) // only trailing newlines trimmed (leading spaces are data)
line := applescript.EscapeLine(text)    // like Escape, plus "\n" → \n and "\r" dropped
```

Failures return `osascript: <stderr>`.

### humanize.Truncate

```go
import "github.com/aeon022/missionctl-core/humanize"

humanize.Truncate("héllo wörld", 6) // "héllo…" — counts runes, ends in "…"
humanize.Truncate("hello", 1)       // "…"
humanize.Truncate("hello", 0)       // "" — safe for n <= 1, unlike s[:n-1]
```

(`humanize.TimeAgo(t)` is the other function in the package.)

### tuitest

Drives Bubble Tea v2 models in tests without a terminal: build the exact
messages a terminal delivers, run them through `Update`, look at the model or
the rendered text. Commands returned by `Update` are collected but **never
executed**. See the [testing tutorial](../go-tutorial/testing-tuis.md).

```go
import (
    "strings"
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/aeon022/missionctl-core/tuitest"
)

func TestSpaceSelects(t *testing.T) {
    var m tea.Model = newModel() // your Bubble Tea v2 model
    m, cmds := tuitest.Send(m, tuitest.Resize(100, 30)) // WindowSizeMsg
    m, cmds = tuitest.Keys(m, "j", "space")             // "enter" "esc" "tab" "shift+tab" "ctrl+c"
                                                        // "up" "down" "pgdown" "f1".."f12" "?" "V" ...
    _ = cmds                                            // returned cmds, not run
    m, _ = tuitest.Send(m, tuitest.Click(3, 5), tuitest.Motion(3, 6), tuitest.Wheel(true, 3, 6))
    if out := tuitest.Text(m); !strings.Contains(out, "selected") { // View().Content, ANSI stripped
        t.Errorf("view:\n%s", out)
    }
}

// Smoke: size the model, press every key, fail on a panic or an empty frame.
func TestSmoke(t *testing.T) {
    tuitest.Smoke(t, newModel(), "j", "k", "?", "esc", "/", "a", "esc")        // 100x30
    tuitest.SmokeSize(t, newModel(), 60, 15, "j", "k", "?", "esc")            // custom size
}
```

`tuitest.Key("space")` carries `Text: " "` like a real terminal; unknown key
names panic so typos fail loudly.

### statusbar

The one-line footer: key hints on the left, status on the right, never wider
than the terminal.

```go
import "github.com/aeon022/missionctl-core/statusbar"

hints := statusbar.Hints(width, // pairs in PRIORITY order: the last ones are dropped first
    [2]string{"?", "help"}, [2]string{"q", "quit"},
    [2]string{"enter", "open"}, [2]string{"x", "action"})
footer := statusbar.Line(width, hints, "synced 2m ago") // left truncated with "…" if needed
```

`statusbar.Hint(key, desc)` renders a single pair.

### emptystate

```go
import "github.com/aeon022/missionctl-core/emptystate"

emptystate.Render(width, height, "✓", "No tasks yet", "press n to add one") // centered block
emptystate.Loading(width, height, m.spinner.View(), "")                     // text defaults to "Loading…"
```

With `width` or `height` ≤ 0, `Render` returns the uncentered block.

### overlay.CenterDim

`Center` with the background dimmed behind the popup, so it reads as modal.
The background loses its own colors (faint plain text): re-coloring styled
lines safely isn't possible.

```go
import "github.com/aeon022/missionctl-core/overlay"

return overlay.CenterDim(m.renderList(), m.renderHelpPopup(), m.width, m.height, 0)
```

### ai hooks

`ai.Detect(prefix)` picks the provider (Anthropic, OpenAI, Gemini, or local
Ollama); `ai.Call` streams the answer. A tool with its own Gemini credentials
(habctl: a Google OAuth refresh token) plugs them in with two optional hooks:

```go
import "github.com/aeon022/missionctl-core/ai"

func init() {
    ai.GeminiAvailable = func() bool { return os.Getenv("GOOGLE_REFRESH_TOKEN") != "" } // cheap, no network
    ai.GeminiKey = func() string { return exchangeForAccessToken() }                   // "" → GEMINI_API_KEY
}
info, err := ai.Detect("HABCTL") // HABCTL_PROVIDER overrides auto-detection
text, err := ai.Call(ctx, info, system, prompt, func(chunk string) { /* stream */ })
```

### syncdir.Acquire creates the data directory

`Acquire(dbPath)` takes the advisory lock on `<dbPath>.lock` and now creates
the database's directory first (`MkdirAll`). Without that, the very first run
on a fresh machine failed with `syncdir: opening …lock: no such file` in
habctl, timectl and diaryctl.

```go
lock, err := syncdir.Acquire(dbPath) // fine even if ~/.local/share/<tool> doesn't exist yet
if errors.Is(err, syncdir.ErrLocked) { /* another process has the DB open */ }
defer lock.Release()
```

### activity

The suite's shared activity log. A tool calls it at the point where the **user** does
something (never in sync/import paths); `missionctl log` and diaryctl read it back.

```go
import "github.com/aeon022/missionctl-core/activity"

activity.Log("taskctl", "completed", task.Title) // best-effort, titles only

from, to := activity.Day(time.Now())
events, _ := activity.Read(from, to)              // oldest first
block := activity.MarkdownBlock(events)           // "<!-- activity:start --> ## Activity …"
body = activity.ReplaceBlock(body, block)         // refreshes the block in place, keeps the rest
```

- **Privacy:** titles only, truncated to 120 runes — never note bodies, mail text,
  recipient addresses or amounts. `Log` never returns an error and never breaks the
  action it describes (a single `O_APPEND` write per line, safe across tools).
- **Where it lives:** `$MISSIONCTL_DATA_DIR` or `~/.local/share/missionctl/activity.jsonl`.
- **Settings** (`~/.config/missionctl/activity.yaml`; `MISSIONCTL_ACTIVITY=off` also disables
  logging): `enabled: true|false` and `diary: ask|auto|off` (what diaryctl does with the
  day's activity). `activity.Load()`, `SetEnabled`, `SetDiaryMode`.
- **Actions** in use: `added`, `completed`, `deleted`, `checked`, `started`, `stopped`,
  `wrote`, `sent`, `unsubscribed`, `published`, `scheduled`, `imported`.
- **Hooking a new intent:** put the call in the lowest function CLI, TUI and MCP share, and
  only log when something actually changed (e.g. habctl logs a check-in only when a row was
  created). Test with `MISSIONCTL_DATA_DIR` and `HOME` pointing at temp dirs.

### ui

The suite's visual vocabulary — small functions returning styled strings built on `theme`, so
every tool follows the user's theme preset and light/dark mode. Preview them all with
`missionctl ui-demo` (`--all` for every preset).

```go
import "github.com/aeon022/missionctl-core/ui"

ui.Pill("overdue", ui.Err)            // filled badge; kinds: Info, OK, Warn, Err, Muted
ui.Hint("enter", "open")              // key cap + dimmed description (ui.KeyCap for the cap alone)
ui.Bar(30, spent/limit, true)         // 1/8-cell bar; true = green→amber→red as it fills
ui.Spark(minutesPerDay)               // ▁▂▃▅█ scaled to the max
ui.Heat(level, 4)                     // one heatmap cell
ui.Toast(ui.OK, "Saved")              // "✓ Saved"
ui.Header(width, "budgetctl", "profile: firma", "synced 2m ago") // drops the middle first
ui.Row(width, selected, text)         // accent bar ▌ + full-width background when selected
ui.RelTime(due, time.Now())           // today · tomorrow · in 3d · 3d ago · Oct 20
ui.MidEllipsis("Rechnung_…_Q3.pdf", 20)
ui.Money(-64.30, 12)                  // right-aligned, red/green, dimmed cents
ui.Icon("mail")                       // ✉ — or the Nerd Font glyph with MISSIONCTL_ICONS=nerd
```

All widths are exact display cells (wide characters included), so the pieces compose without
wrapping. Nerd Font glyphs are opt-in; the default needs no special font.

### keymap.Text

`Help.Text(line)` adds one freeform line to the current section — for
explanations that don't fit the key/description shape. (It was once removed as
"unused"; budgetctl's help calls it.)

```go
keymap.Bare().
    Section("Accounts").
    Text("An account is just a tag on a transaction.").
    String()
```

## Roadmap

- [x] Shared theme (`theme` package)
- [x] Shared help-overlay builder (`keymap` package)
- [x] Shared standard keymap constants (`/` search, `?` help, `d` delete+confirm, `q`/`esc` quit)
- [x] Shared spinner constructor (`theme.NewSpinner`)
- [x] Data dir helpers (`config.DataDir` / `config.ResolveDir`)
- [x] Safe cross-device data-dir sync helpers (`syncdir` + `config.ResolveDir`)
- [x] `config.Store` (YAML + opt-in env prefix), replacing viper in four tools
- [x] `applescript`, `humanize.Truncate` (de-duplicated from several tools)
- [x] `tuitest` (test driver + smoke tests), `statusbar`, `emptystate`, `overlay.CenterDim`
- [x] `ai` Gemini credential hooks; `syncdir.Acquire` creates the data directory
- [ ] License-check helper — deliberately not started: monetization (and
  therefore the license model it would check against) hasn't been decided
  yet, see `MONETIZATION.md` and `ROADMAP.md` in the root repo
