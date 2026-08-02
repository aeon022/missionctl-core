// Package syncdir makes it safe for a tool to point its SQLite database at
// a folder the user syncs themselves (iCloud Drive, Dropbox, Syncthing,
// ...) instead of the tool's private default directory. Three separate
// hazards show up the moment a database file lives in a folder like that,
// and this package exists because none of them are solved by SQLite or the
// sync client on their own:
//
//  1. WAL journal mode splits the true on-disk state across up to three
//     files (name.db, name.db-wal, name.db-shm). A folder-sync client has
//     no idea those three files must be captured together — it uploads
//     whichever one changed, whenever it changed, with no cross-file
//     atomicity. That's a recipe for a device seeing a .db file whose
//     .db-wal hasn't caught up yet, or the reverse. JournalMode switches a
//     tool to the classic rollback-journal mode instead the moment the
//     tool is pointed at a user-configured (potentially synced) directory:
//     rollback-journal's on-disk footprint is the single main file plus a
//     transient "-journal" sidecar that exists only during an active write
//     transaction and is deleted immediately on commit, so whenever the
//     app isn't actively mid-write — which is the only time a sync client
//     can realistically take a consistent snapshot anyway — the directory
//     holds exactly one file. The tool's private default directory is left
//     on WAL, unchanged: this only applies once a user opts in.
//  2. Two processes on the same machine (two terminal tabs, a crashed
//     session that never released its handle) must not both hold the file
//     open for writing at once. Acquire/Release is a same-machine advisory
//     lock (flock(2), LOCK_EX|LOCK_NB) for exactly that case. It needs no
//     crash or signal handling to stay correct: the kernel releases a
//     flock automatically the moment the holding process exits, however it
//     exits.
//  3. macOS's "Optimize Mac Storage" can evict an iCloud Drive file to save
//     local disk space, replacing it with a zero-byte ".name.icloud"
//     placeholder until something downloads it back on demand. Without
//     detecting this, a tool just sees a missing file (or silently starts
//     a fresh empty database) with no indication why. ICloudPlaceholder
//     catches that specific case so a tool can give an actionable message.
//
// None of this makes two live machines editing the same row at the same
// instant safe — that would need a real distributed lock or a conflict-
// resolution protocol, and no folder-sync client offers the former. This
// package only de-risks the much more common failure modes: a stale
// multi-file WAL snapshot, an accidental double-open on one machine, and
// an undownloaded placeholder file.
package syncdir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// JournalMode returns the SQLite journal-mode DSN value a tool should use
// for its database connection. WAL is left untouched for a tool's private
// default directory (shared == false); the moment a user points the tool
// at a configured directory of their own, shared == true switches to DELETE
// (rollback-journal) — see the package doc comment for why.
func JournalMode(shared bool) string {
	if shared {
		return "DELETE"
	}
	return "WAL"
}

// ErrLocked is returned by Acquire when another process already holds the
// lock on this database file.
var ErrLocked = errors.New("syncdir: database is already open in another process")

// Lock is an advisory, same-machine-only exclusive lock on a database
// file, meant to be held for the whole lifetime of the process that
// acquired it.
type Lock struct {
	f *os.File
}

// Acquire non-blockingly takes an exclusive lock on "<dbPath>.lock",
// creating that file if it doesn't exist yet. It returns ErrLocked
// (wrapped) if another live process already holds it.
func Acquire(dbPath string) (*Lock, error) {
	lockPath := dbPath + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("syncdir: opening %s: %w", lockPath, err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		if err == unix.EWOULDBLOCK {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("syncdir: locking %s: %w", lockPath, err)
	}
	return &Lock{f: f}, nil
}

// Release releases the lock and closes its file descriptor. It's safe to
// call on a nil *Lock (a no-op), so callers can defer it unconditionally
// even on the error path of Acquire.
func (l *Lock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	err := unix.Flock(int(l.f.Fd()), unix.LOCK_UN)
	if cerr := l.f.Close(); err == nil {
		err = cerr
	}
	return err
}

// ICloudPlaceholder reports whether dbPath is currently an undownloaded
// iCloud Drive stub: dbPath itself doesn't exist, but a sibling
// ".<name>.icloud" file does. placeholderPath is that sibling's path, for
// use in an error message. If dbPath exists, or neither file exists (the
// database just hasn't been created yet), it returns false.
func ICloudPlaceholder(dbPath string) (isPlaceholder bool, placeholderPath string) {
	if _, err := os.Stat(dbPath); err == nil {
		return false, ""
	}
	dir, name := filepath.Split(dbPath)
	candidate := filepath.Join(dir, "."+name+".icloud")
	if _, err := os.Stat(candidate); err == nil {
		return true, candidate
	}
	return false, ""
}
