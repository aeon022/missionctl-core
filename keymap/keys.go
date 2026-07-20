// Keys the suite's TUIs already agree on by convention (verified against
// mailctl, calctl, taskctl, notectl, budgetctl, habctl, timectl, diaryctl):
// "/" opens search, "?" toggles help, "d" asks to delete, "y"/enter confirms,
// "q"/"esc" quits or backs out of the current view. habctl's "s" is AI-suggest
// rather than sync since it has no external data source to sync from — that's
// a deliberate exception, not drift.
package keymap

const (
	SearchKey  = "/"
	HelpKey    = "?"
	DeleteKey  = "d"
	ConfirmKey = "y"
	QuitKey    = "q"
	BackKey    = "esc"
	SyncKey    = "s"
)
