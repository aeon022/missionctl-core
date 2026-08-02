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
  DB on purpose; this package doesn't override that). `ResolveDir(tool,
  override)` returns `override` (expanded and created) if the user
  configured one, else falls back to `DataDir(tool)` — the entry point for
  pointing a tool's data at a folder of the user's own choosing, see
  `syncdir` below.
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
  was added.

## Adoption

Migration is **incremental, not a big-bang rewrite**: a tool adopts
`missionctl-core` the next time its TUI is touched for another reason, by
adding the dependency and replacing its local color vars / help-overlay
helpers with calls into this package. No tool is required to migrate on any
particular timeline.

Status: all seven target tools now depend on `theme` for the color palette.
`keymap` (help-overlay builder) and `config` (dir helpers) are still
per-tool-copied and not yet migrated — same incremental policy applies.

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
- [x] Safe cross-device data-dir sync helpers (`syncdir` + `config.ResolveDir`)
- [ ] License-check helper — deliberately not started: monetization (and
  therefore the license model it would check against) hasn't been decided
  yet, see `MONETIZATION.md` and `ROADMAP.md` in the root repo
