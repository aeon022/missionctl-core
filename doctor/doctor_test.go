package doctor

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestCheckFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if c := CheckFileExists("config", path); c.OK {
		t.Fatal("expected missing file to fail")
	}

	if err := os.WriteFile(path, []byte("x: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if c := CheckFileExists("config", path); !c.OK {
		t.Fatalf("expected existing file to pass, got %+v", c)
	}
}

func TestCheckSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	if c := CheckSQLite("db", dbPath, "items"); c.OK {
		t.Fatal("expected missing db to fail")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO items DEFAULT VALUES"); err != nil {
		t.Fatal(err)
	}
	db.Close()

	c := CheckSQLite("db", dbPath, "items")
	if !c.OK {
		t.Fatalf("expected valid db+table to pass, got %+v", c)
	}

	c = CheckSQLite("db", dbPath, "nonexistent_table")
	if c.OK {
		t.Fatal("expected missing table to fail")
	}
}

func TestCheckAppleApp(t *testing.T) {
	if c := CheckAppleApp("Calendar", "Calendar"); !c.OK {
		t.Fatalf("expected a real installed app to pass, got %+v", c)
	}
	if c := CheckAppleApp("Fake", "ThisAppDoesNotExist12345"); c.OK {
		t.Fatal("expected a nonexistent app name to fail, not report reachable")
	}
}

func TestCheckDataDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	if err := os.WriteFile(dbPath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := CheckDataDir("Data directory", dbPath, false)
	if !c.OK {
		t.Fatalf("expected local mode to pass, got %+v", c)
	}
	if !strings.Contains(c.Detail, "(local)") {
		t.Errorf("expected detail to mention local mode, got %q", c.Detail)
	}

	c = CheckDataDir("Data directory", dbPath, true)
	if !c.OK || !strings.Contains(c.Detail, "(shared)") {
		t.Errorf("expected shared mode to pass and mention shared, got %+v", c)
	}
}

func TestCheckDataDirICloudPlaceholder(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	placeholder := filepath.Join(dir, ".test.db.icloud")
	if err := os.WriteFile(placeholder, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	c := CheckDataDir("Data directory", dbPath, true)
	if c.OK {
		t.Fatal("expected an undownloaded iCloud placeholder to fail")
	}
	if !strings.Contains(c.Detail, "iCloud") {
		t.Errorf("expected detail to explain the iCloud placeholder, got %q", c.Detail)
	}
}

func TestPrintReport(t *testing.T) {
	if ok := PrintReport([]Check{{Label: "a", OK: true}}); !ok {
		t.Fatal("expected all-passing report to return true")
	}
	if ok := PrintReport([]Check{{Label: "a", OK: true}, {Label: "b", OK: false}}); ok {
		t.Fatal("expected a report with one failure to return false")
	}
}
