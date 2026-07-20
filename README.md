# missionctl-core

Shared Bubble Tea / Lipgloss building blocks for the [missionctl](https://github.com/aeon022/missionctl) suite.

## Why this exists

mailctl, calctl, taskctl, notectl, budgetctl and timectl had already converged
on an identical color palette and near-identical help-overlay code by
copy-paste. This module makes that convergence explicit and shared instead of
accidental and duplicated. habctl keeps its own distinct palette on purpose
and is not expected to depend on this package.

## Packages

- `theme` — the shared `AdaptiveColor` palette (Blue/Green/Red/Amber/Muted/
  Subtle) and base Lipgloss styles used across the suite's TUIs, plus
  `NewSpinner` for the MiniDot sync/AI-call spinner every tool already builds
  the same way.
- `keymap` — a small builder for the `?` help overlay (replacing each tool's
  hand-rolled `key()/row()/section()` helpers), plus the standard single-key
  bindings (`SearchKey "/"`, `HelpKey "?"`, `DeleteKey "d"`, `ConfirmKey "y"`,
  `QuitKey "q"`, `BackKey "esc"`, `SyncKey "s"`) already followed by every
  tool except habctl's `s` (AI-suggest, not sync — deliberate, habctl has no
  external data source to sync from).
- `config` — `Dir(tool)` / `DataDir(tool)` for the `~/.config/<tool>` and
  `~/.local/share/<tool>` conventions most tools already follow (mailctl and
  taskctl keep using `~/Library/Application Support/<tool>` for their SQLite
  DB on purpose; this package doesn't override that).

## Adoption

Migration is **incremental, not a big-bang rewrite**: a tool adopts
`missionctl-core` the next time its TUI is touched for another reason, by
adding the dependency and replacing its local color vars / help-overlay
helpers with calls into this package. No tool is required to migrate on any
particular timeline.

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

## Roadmap

- [x] Shared theme (`theme` package)
- [x] Shared help-overlay builder (`keymap` package)
- [x] Shared standard keymap constants (`/` search, `?` help, `d` delete+confirm, `q`/`esc` quit)
- [x] Shared spinner constructor (`theme.NewSpinner`)
- [x] Config/data dir helpers (`config.Dir` / `config.DataDir`)
- [ ] License-check helper — deliberately not started: monetization (and
  therefore the license model it would check against) hasn't been decided
  yet, see `MONETIZATION.md` and `ROADMAP.md` in the root repo
