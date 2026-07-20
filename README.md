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
  Subtle) and base Lipgloss styles used across the suite's TUIs.
- `keymap` — a small builder for the `?` help overlay, replacing each tool's
  hand-rolled `key()/row()/section()` helpers with one implementation.

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
- [ ] Shared standard keymap constants (`/` search, `?` help, `d` delete+confirm, `q`/`esc` quit)
- [ ] Shared spinner/sync `tea.Cmd` pattern
- [ ] Config helpers (`~/.config/missionctl/`)
- [ ] License-check helper (once monetization starts — see `MONETIZATION.md` in the root repo)
